package cloudconfig

const DecryptTLSAssetsScript = `#!/bin/bash -e
{{ if eq .EncrypterType "vault" }}
token_path=/var/token
nonce_path=/var/nonce

wait_for_vault_elb(){
    local service_name="$1"
    local state="${2:-active}"
    for i in $(seq 80); do
        if curl -k -s -o /dev/null -w "%{http_code}" --max-time 3 {{ .VaultAddress }}/v1/sys/health | grep -q "200"; then
            return 0
        fi
        echo "{{ .VaultAddress }} not accessible yet, waiting..."
        sleep 15;
    done

    echo "{{ .VaultAddress }} not accessible"
    return 1
}

token_exists () {
  if [ -f $token_path ]; then
    return 0
  else
    return 1
  fi
}

token_is_valid() {
  #  https://www.vaultproject.io/api/auth/token/index.html#lookup-a-token-self-
  echo "Checking token validity"
  token_lookup=$(curl -k \
    --request GET \
    --silent \
    --header "X-Vault-Token: $(cat $token_path)" \
    --write-out %{http_code} \
    --output /dev/null \
    {{ .VaultAddress }}/v1/auth/token/lookup-self)
  if [ "$token_lookup" == "200" ]; then
      echo "$0 - Valid token found, exiting"
      return 0
  else
      echo "$0 - Invalid token found"
      return 1
  fi
}

main () {
   if ! wait_for_vault_elb; then
        exit 1;
   fi
   if ! token_exists; then
        aws_login ""
    elif token_exists && ! token_is_valid; then
        aws_login "$(cat $nonce_path)"
    else
        logger $0 "current vault token is still valid"
    fi

    echo decrypting keys assets
    shopt -s nullglob
    for encKey in $(find /etc/kubernetes/ssl -name "*.pem.enc"); do
      echo decrypting $encKey
      f=$(mktemp $encKey.XXXXXXXX)
      cat <<EOF > data.json
{
  "ciphertext": "$(cat $encKey)"
}
EOF
      curl -k \
        --header "X-Vault-Token: $(cat $token_path)" \
        --silent \
        --request POST \
        --data @data.json \
        {{ .VaultAddress }}/v1/transit/decrypt/{{ .EncryptionKey }} | \
        jq -r .data.plaintext | base64 -d > $f
      mv -f $f ${encKey%.enc}
      rm -f data.json
    done;
    echo done.
}

aws_login () {
    # query EC2 metadata endpoint (common for all AWS infrastructure).
    pkcs7=$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/pkcs7 | tr -d '\n')
    if [ -z "$1" ]; then
        # do not load nonce if initial login
        login_payload=$(cat <<EOF
{
  "role": "decrypter",
  "pkcs7": "$pkcs7"
}
EOF
)
    else
        # load nonce in payload for reauthentication
        login_payload=$(cat <<EOF
{
  "role": "decrypter",
  "pkcs7": "$pkcs7",
  "nonce": "$1"
}
EOF
)
    fi

    curl -k \
      --request POST \
      --silent \
      --data "$login_payload" \
      {{ .VaultAddress }}/v1/auth/aws/login | tee  \
      >(jq -r .auth.client_token > $token_path) \
      >(jq -r .auth.metadata.nonce > $nonce_path)
}

main
{{ else }}
rkt run \
  --volume=ssl,kind=host,source=/etc/kubernetes/ssl,readOnly=false \
  --mount=volume=ssl,target=/etc/kubernetes/ssl \
  --uuid-file-save=/var/run/coreos/decrypt-tls-assets.uuid \
  --volume=dns,kind=host,source=/etc/resolv.conf,readOnly=true --mount volume=dns,target=/etc/resolv.conf \
  --net=host \
  --trust-keys-from-https \
  quay.io/coreos/awscli:025a357f05242fdad6a81e8a6b520098aa65a600 --exec=/bin/bash -- \
    -ec \
    'echo decrypting tls assets
    shopt -s nullglob
    for encKey in $(find /etc/kubernetes/ssl -name "*.pem.enc"); do
      echo decrypting $encKey
      f=$(mktemp $encKey.XXXXXXXX)
      /usr/bin/aws \
        --region {{.AWS.Region}} kms decrypt \
        --ciphertext-blob fileb://$encKey \
        --output text \
        --query Plaintext \
      | base64 -d > $f
      mv -f $f ${encKey%.enc}
    done;'

rkt rm --uuid-file=/var/run/coreos/decrypt-tls-assets.uuid || :
{{ end }}
chown -R etcd:etcd /etc/kubernetes/ssl/etcd`

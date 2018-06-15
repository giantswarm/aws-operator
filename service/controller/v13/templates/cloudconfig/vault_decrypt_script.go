package cloudconfig

const VaultDecryptScript = `
token_path=/var/token
nonce_path=/var/nonce

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
    if ! token_exists; then
        aws_login ""
    elif token_exists && ! token_is_valid; then
        aws_login "$(cat $nonce_path)"
    else
        logger $0 "current vault token is still valid"
    fi

    echo decrypting keys assets
    shopt -s nullglob
    for encKey in $(find /etc/kubernetes/encryption -name "*.enc"); do
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
`

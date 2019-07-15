package cloudconfig

const DecryptTLSAssetsScript = `#!/bin/bash -e

token_path=/var/token

vault_tls_assets_decrypt() {
    echo decrypting tls assets
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

kms_tls_assets_decrypt() {
AWS_CLI_IMAGE="quay.io/coreos/awscli:025a357f05242fdad6a81e8a6b520098aa65a600"

while ! rkt fetch --trust-keys-from-https=true ${AWS_CLI_IMAGE};
do
        echo "Failed to fetch rkt image ${AWS_CLI_IMAGE}, retrying in 5 sec."
        sleep 5s
done
echo "Successfully fetched rkt image ${AWS_CLI_IMAGE}."

rkt run \
  --volume=ssl,kind=host,source=/etc/kubernetes/ssl,readOnly=false \
  --mount=volume=ssl,target=/etc/kubernetes/ssl \
  --uuid-file-save=/var/run/coreos/decrypt-tls-assets.uuid \
  --volume=dns,kind=host,source=/etc/resolv.conf,readOnly=true --mount volume=dns,target=/etc/resolv.conf \
  --net=host \
  --trust-keys-from-https \
  ${AWS_CLI_IMAGE} --exec=/bin/bash -- \
    -ec \
    'echo decrypting tls assets
    shopt -s nullglob
    for encKey in $(find /etc/kubernetes/ssl -name "*.pem.enc"); do
      echo decrypting $encKey
      f=$(mktemp $encKey.XXXXXXXX)
      /usr/bin/aws \
        --region {{.AWSRegion}} kms decrypt \
        --ciphertext-blob fileb://$encKey \
        --output text \
        --query Plaintext \
      | base64 -d > $f
      mv -f $f ${encKey%.enc}
    done;'

rkt rm --uuid-file=/var/run/coreos/decrypt-tls-assets.uuid || :
}


main() {
{{ if eq .EncrypterType "vault" }}
  vault_tls_assets_decrypt
{{ else }}
  kms_tls_assets_decrypt
{{ end }}

chown -R etcd:etcd /etc/kubernetes/ssl/etcd
}

main
`

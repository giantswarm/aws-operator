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
AWS_CLI_IMAGE="{{.RegistryDomain}}/giantswarm/awscli:1.18.3"

while ! docker pull ${AWS_CLI_IMAGE};
do
        echo "Failed to fetch docker image ${AWS_CLI_IMAGE}, retrying in 5 sec."
        sleep 5s
done
echo "Successfully fetched docker image ${AWS_CLI_IMAGE}."


docker run --net=host -v /etc/kubernetes/ssl:/etc/kubernetes/ssl \
        --entrypoint=/bin/sh \
        ${AWS_CLI_IMAGE} \
        -ec \
        'echo decrypting tls assets
    for encKey in $(find /etc/kubernetes/ssl -name "*.pem.enc"); do
      echo decrypting $encKey
      f=$(mktemp $encKey.XXXXXXXX)
      aws \
        --region {{.AWSRegion}} kms decrypt \
        --ciphertext-blob fileb://$encKey \
        --output text \
        --query Plaintext \
      | base64 -d > $f
      mv -f $f ${encKey%.enc}
    done;'
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

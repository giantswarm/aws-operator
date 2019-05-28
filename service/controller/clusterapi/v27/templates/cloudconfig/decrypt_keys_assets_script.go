package cloudconfig

const DecryptKeysAssetsScript = `#!/bin/bash -e

token_path=/var/token

vault_keys_assets_decrypt() {
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

kms_keys_assets_decrypt() {
rkt run \
  --volume=keys,kind=host,source=/etc/kubernetes/encryption,readOnly=false \
  --mount=volume=keys,target=/etc/kubernetes/encryption \
  --uuid-file-save=/var/run/coreos/decrypt-keys-assets.uuid \
  --volume=dns,kind=host,source=/etc/resolv.conf,readOnly=true --mount volume=dns,target=/etc/resolv.conf \
  --net=host \
  --trust-keys-from-https \
  quay.io/coreos/awscli:025a357f05242fdad6a81e8a6b520098aa65a600 --exec=/bin/bash -- \
    -ec \
    'echo decrypting keys assets
    shopt -s nullglob
    for encKey in $(find /etc/kubernetes/encryption -name "*.enc"); do
      echo decrypting $encKey
      f=$(mktemp $encKey.XXXXXXXX)
      /usr/bin/aws \
        --region {{.AWSRegion}} kms decrypt \
        --ciphertext-blob fileb://$encKey \
        --output text \
        --query Plaintext \
      | base64 -d > $f
      mv -f $f ${encKey%.enc}
    done;
    echo done.'

rkt rm --uuid-file=/var/run/coreos/decrypt-keys-assets.uuid || :
}

main() {
{{ if eq .EncrypterType "vault" }}
  vault_keys_assets_decrypt
{{ else }}
  kms_keys_assets_decrypt
{{ end }}

chown -R etcd:etcd /etc/kubernetes/ssl/etcd
}

main
`

package cloudconfig

const DecryptKeysAssetsScript = `#!/bin/bash -e
{{ if eq .EncrypterType "vault" }}
{{ .VaultDecryptScript  }}
{{ else }}
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
        --region {{.AWS.Region}} kms decrypt \
        --ciphertext-blob fileb://$encKey \
        --output text \
        --query Plaintext \
      | base64 -d > $f
      mv -f $f ${encKey%.enc}
    done;
    echo done.'

rkt rm --uuid-file=/var/run/coreos/decrypt-keys-assets.uuid || :
{{ end }}
`

package cloudconfig

const DecryptTLSAssetsScript = `#!/bin/bash -e

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

chown -R etcd:etcd /etc/kubernetes/ssl/etcd`

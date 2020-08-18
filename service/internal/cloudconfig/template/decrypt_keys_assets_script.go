package template

const DecryptKeysAssetsScript = `#!/bin/bash -e
set -o errexit

kms_keys_assets_decrypt() {
AWS_CLI_IMAGE="{{.RegistryDomain}}/giantswarm/awscli:1.18.3"

while ! docker pull ${AWS_CLI_IMAGE};
do
        echo "Failed to fetch docker image ${AWS_CLI_IMAGE}, retrying in 5 sec."
        sleep 5s
done
echo "Successfully fetched docker image ${AWS_CLI_IMAGE}."


docker run --net=host -v /etc/kubernetes/encryption:/etc/kubernetes/encryption \
        --entrypoint=/bin/sh \
        ${AWS_CLI_IMAGE} \
        -ec \
        'set -o errexit
    echo decrypting tls assets
    for encKey in $(find /etc/kubernetes/encryption -name "*.enc"); do
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
  kms_keys_assets_decrypt
  chown -R etcd:etcd /etc/kubernetes/ssl/etcd
}

main
`

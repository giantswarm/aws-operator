package template

const DecryptTLSAssetsScript = `#!/bin/bash -e
set -o errexit

kms_tls_assets_decrypt() {
AWS_CLI_IMAGE="{{.RegistryDomain}}/giantswarm/awscli:1.18.3"

while ! docker pull ${AWS_CLI_IMAGE};
do
        echo "Failed to fetch docker image ${AWS_CLI_IMAGE}, retrying in 5 sec."
        sleep 5s
done
echo "Successfully fetched docker image ${AWS_CLI_IMAGE}."


while ! docker run --net=host -v /etc/kubernetes/ssl:/etc/kubernetes/ssl \
        --entrypoint=/bin/sh \
        ${AWS_CLI_IMAGE} \
        -ec \
        'set -o errexit
    echo decrypting tls assets
    for encKey in $(find /etc/kubernetes/ssl -name "*.pem.enc"); do
      echo decrypting $encKey
      f=$(mktemp $encKeyb64.XXXXXXXX)
      f2=$(mktemp $encKey.XXXXXXXX)
      aws \
        --region {{.AWSRegion}} kms decrypt \
        --ciphertext-blob fileb://$encKey \
        --output text \
        --query Plaintext > $f
      base64 -d $f > $f2
      mv -f $f2 ${encKey%.enc}
    done;'
do
		echo "Failed to decrypt TLS assets, retrying in 5 sec."
        sleep 5s
done

}


main() {
  kms_tls_assets_decrypt
  if [ -d "/etc/kubernetes/ssl/etcd" ]; then 
    chown -R etcd:etcd /etc/kubernetes/ssl/etcd
  fi
}

main
`

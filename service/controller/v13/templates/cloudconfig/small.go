package cloudconfig

const Small = `#!/bin/bash

# user-data in EC2 instances has a 16KB limit.
# To circumvent this limit, we:
#
# 1. Upload the final cloudconfig to s3
# 2. Generate a "small cloudconfig" whose only task is fetching the
#    final cloudconfig from s3
# 3. Configure the instance to be able to access the s3 URI where the
#    final cloudconfig is stored
# 4. Start the instance with the "small cloudconfig"
#
# This file is the "small cloudconfig" mentioned before. Here we simply fetch a
# gzip+base64 file (the final cloudconfig) from AWS S3 and run coreos-cloudinit
# with it as an argument.

. /etc/environment
USERDATA_FILE={{.MachineType}}

# Wait for S3 bucket to be available.
s3_http_uri="https://{{.S3Domain}}/{{.S3URI}}/cloudconfig/{{.CloudConfigVersion}}/$USERDATA_FILE"
retry=30

until [ $(curl --output /dev/null --silent --head --fail -w "%{http_code}" $s3_http_uri) -eq "403" ]; do
   retry=$(( retry - 1))
   if [ $retry -le 0 ]; then
     echo "timed out waiting for s3 bucket"
     exit 1
   fi

   echo "waiting for $s3_http_uri to be available"
   sleep 5
done

/usr/bin/rkt run \
    --net=host \
    --volume=dns,kind=host,source=/etc/resolv.conf,readOnly=true --mount volume=dns,target=/etc/resolv.conf  \
    --volume=awsenv,kind=host,source=/var/run/coreos,readOnly=false --mount volume=awsenv,target=/var/run/coreos \
    --trust-keys-from-https \
    --insecure-options=image docker://registry-intl.cn-shanghai.aliyuncs.com/giantswarm/awscli:latest -- aws s3 --region {{.Region}} cp s3://{{.S3URI}}/cloudconfig/{{.CloudConfigVersion}}/$USERDATA_FILE /var/run/coreos/temp.txt
base64 -d /var/run/coreos/temp.txt | gunzip > /var/run/coreos/$USERDATA_FILE
exec /usr/bin/coreos-cloudinit --from-file /var/run/coreos/$USERDATA_FILE`

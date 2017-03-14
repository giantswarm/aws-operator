#!/bin/bash

. /etc/environment
USERDATA_FILE={{.MachineType}}

/usr/bin/rkt run \
    --net=host \
    --volume=dns,kind=host,source=/etc/resolv.conf,readOnly=true --mount volume=dns,target=/etc/resolv.conf  \
    --volume=awsenv,kind=host,source=/var/run/coreos,readOnly=false --mount volume=awsenv,target=/var/run/coreos \
    --trust-keys-from-https \
    quay.io/coreos/awscli:025a357f05242fdad6a81e8a6b520098aa65a600 -- aws s3 --region {{.Region}} cp s3://{{.S3DirURI}}/$USERDATA_FILE /var/run/coreos/temp.txt
base64 -d /var/run/coreos/temp.txt | gunzip > /var/run/coreos/$USERDATA_FILE
exec /usr/bin/coreos-cloudinit --from-file /var/run/coreos/$USERDATA_FILE
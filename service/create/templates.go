package create

const (
	decryptTLSAssetsScriptTemplate = `#!/bin/bash -e

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
		done;
		echo done.'

rkt rm --uuid-file=/var/run/coreos/decrypt-tls-assets.uuid || :

chown -R etcd:etcd /etc/kubernetes/ssl/etcd`

	decryptTLSAssetsServiceTemplate = `
[Unit]
Description=Decrypt TLS certificates

[Service]
Type=oneshot
ExecStart=/opt/bin/decrypt-tls-assets

[Install]
WantedBy=multi-user.target
`

	masterFormatVarLibDockerServiceTemplate = `
[Unit]
Description=Format /var/lib/docker to XFS
Before=docker.service var-lib-docker.mount
ConditionPathExists=!/var/lib/docker

[Service]
Type=oneshot
ExecStart=/usr/sbin/mkfs.xfs -f /dev/xvdb

[Install]
WantedBy=multi-user.target
`

	workerFormatVarLibDockerServiceTemplate = `
[Unit]
Description=Format /var/lib/docker to XFS
Before=docker.service var-lib-docker.mount
ConditionPathExists=!/var/lib/docker

[Service]
Type=oneshot
ExecStart=/usr/sbin/mkfs.xfs -f /dev/xvdh

[Install]
WantedBy=multi-user.target
`

	ephemeralVarLibDockerMountTemplate = `
[Unit]
Description=Mount ephemeral volume on /var/lib/docker

[Mount]
What=/dev/xvdb
Where=/var/lib/docker
Type=xfs

[Install]
RequiredBy=local-fs.target
`
	persistentVarLibDockerMountTemplate = `
[Unit]
Description=Mount persistent volume on /var/lib/docker

[Mount]
What=/dev/xvdh
Where=/var/lib/docker
Type=xfs

[Install]
RequiredBy=local-fs.target
`

	waitDockerConfTemplate = `
[Unit]
After=var-lib-docker.mount
Requires=var-lib-docker.mount
`

	instanceStorageTemplate = `
storage:
  filesystems:
    - name: ephemeral1
      mount:
        device: /dev/xvdb
        format: ext3
        create:
          force: true
`

	userDataScriptTemplate = `#!/bin/bash

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

/usr/bin/rkt run \
    --net=host \
    --volume=dns,kind=host,source=/etc/resolv.conf,readOnly=true --mount volume=dns,target=/etc/resolv.conf  \
    --volume=awsenv,kind=host,source=/var/run/coreos,readOnly=false --mount volume=awsenv,target=/var/run/coreos \
    --trust-keys-from-https \
    quay.io/coreos/awscli:025a357f05242fdad6a81e8a6b520098aa65a600 -- aws s3 --region {{.Region}} cp s3://{{.S3DirURI}}/$USERDATA_FILE /var/run/coreos/temp.txt
base64 -d /var/run/coreos/temp.txt | gunzip > /var/run/coreos/$USERDATA_FILE
exec /usr/bin/coreos-cloudinit --from-file /var/run/coreos/$USERDATA_FILE`
)

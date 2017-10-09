package cloudconfig

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

	instanceStorageClassTemplate = `
write_files:
- path: /srv/default-storage-class.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: storage.k8s.io/v1beta1
    kind: StorageClass
    metadata:
      name: gp2
      annotations:
        storageclass.beta.kubernetes.io/is-default-class: "true"
      labels:
        kubernetes.io/cluster-service: "true"
        addonmanager.kubernetes.io/mode: EnsureExists
    provisioner: kubernetes.io/aws-ebs
    parameters:
      type: gp2
      encrypted: "true" 
`

	encryptionConfigTemplate = `
kind: EncryptionConfig
apiVersion: v1
resources:
  - resources:
    - secrets
    providers:
    - identity: {}
    - aescbc:
        keys:
        - name: key1
          secret: {{.Cluster.Kubernetes.EncryptionKey}}
`
)

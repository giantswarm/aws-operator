package cloudconfig

const (
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
        format: xfs
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

	ingressControllerConfigMapTemplate = `kind: ConfigMap
apiVersion: v1
metadata:
  name: ingress-nginx
  namespace: kube-system
  labels:
    k8s-addon: ingress-nginx.addons.k8s.io
data:
  server-name-hash-bucket-size: "1024"
  server-name-hash-max-size: "1024"
  use-proxy-protocol: "true"
`
	formatEtcdVolume = `
[Unit]
Description=Formats EBS /dev/xvdh volume
Requires=dev-xvdh.device
After=dev-xvdh.device
ConditionPathExists=!/var/lib/etcd-volume-formated

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/sbin/mkfs.ext4 /dev/xvdh
ExecStartPost=/usr/bin/touch /var/lib/etcd-volume-formated

[Install]
WantedBy=multi-user.target
`
	mountEtcdVolume = `
[Unit]
Description=etcd3 data volume
Requires=format-etcd-ebs.service
After=format-etcd-ebs.service
Before=set-ownership-etcd-data-dir.service etcd3.service

[Mount]
What=/dev/xvdh
Where=/etc/kubernetes/data/etcd
Type=ext4

[Install]
WantedBy=multi-user.target
`
)

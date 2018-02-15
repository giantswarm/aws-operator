package cloudconfig

const (
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

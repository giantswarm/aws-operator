package cloudconfig

const (
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

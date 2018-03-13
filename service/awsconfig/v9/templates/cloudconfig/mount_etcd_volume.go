package cloudconfig

const MountEtcdVolume = `
[Unit]
Description=etcd3 data volume
Requires=format-etcd-ebs.service
After=format-etcd-ebs.service
Before=etcd3.service

[Mount]
What=/dev/xvdh
Where=/var/lib/etcd
Type=ext4

[Install]
WantedBy=multi-user.target
`

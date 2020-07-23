package template

const MountEtcdVolumeAsgMasters = `
[Unit]
Description=etcd3 data volume
Before=etcd3.service

[Mount]
What=/dev/disk/by-label/etcd
Where=/var/lib/etcd
Type=ext4
DirectoryMode=0700

[Install]
WantedBy=multi-user.target
`

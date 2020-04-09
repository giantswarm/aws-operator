package template

const AutomountEtcdVolume = `
[Unit]
Description=etcd3 data volume
After=etcd3-attach-dependencies.service
Before=etcd3.service

[Mount]
Where=/var/lib/etcd

[Install]
WantedBy=multi-user.target
`

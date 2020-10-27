package template

const AutomountEtcdVolume = `
[Unit]
Description=etcd3 data volume

[Mount]
Where=/var/lib/etcd

[Install]
WantedBy=multi-user.target
`

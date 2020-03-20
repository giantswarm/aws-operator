package cloudconfig

const Etcd3ExtraConfig = `
[Unit]
Requires=etcd3-attach-dependencies.service
After=etcd3-attach-dependencies.service
`

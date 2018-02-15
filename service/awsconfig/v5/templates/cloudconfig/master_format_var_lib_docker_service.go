package cloudconfig

const MasterFormatVarLibDockerServiceTemplate = `
[Unit]
Description=Format /var/lib/docker to XFS
Before=docker.service var-lib-docker.mount

[Service]
Type=oneshot
ExecStart=/usr/sbin/mkfs.xfs -f /dev/xvdb

[Install]
WantedBy=multi-user.target
`

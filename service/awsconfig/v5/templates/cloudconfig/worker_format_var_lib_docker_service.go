package cloudconfig

const WorkerFormatVarLibDockerServiceTemplate = `
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

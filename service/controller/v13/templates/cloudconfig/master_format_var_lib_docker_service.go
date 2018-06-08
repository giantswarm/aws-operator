package cloudconfig

const MasterFormatVarLibDockerService = `
[Unit]
Description=Format /var/lib/docker to XFS
Before=docker.service var-lib-docker.mount
ConditionPathExists=!/var/lib/docker

[Service]
Type=oneshot
ExecStart=/bin/bash -c "[ -e "/dev/xvdc" ] && /usr/sbin/mkfs.xfs -f /dev/xvdc -L docker"

[Install]
WantedBy=multi-user.target
`

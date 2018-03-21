package cloudconfig

const WorkerFormatVarLibDockerService = `
[Unit]
Description=Format /var/lib/docker to XFS
Before=docker.service var-lib-docker.mount
ConditionPathExists=!/var/lib/docker

[Service]
Type=oneshot
ExecStart=/bin/bash -c "[ -b "/dev/xvdh" ] && /usr/sbin/mkfs.xfs -f /dev/xvdh -L docker || [ -b "/dev/nvme1n1" ] && /usr/sbin/mkfs.xfs -f /dev/nvme1n1 -L docker"

[Install]
WantedBy=multi-user.target
`

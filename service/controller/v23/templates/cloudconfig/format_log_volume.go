package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
Before=docker.service var-log.mount

[Service]
Type=oneshot
RemainAfterExit=yes

ExecStart=/bin/bash -c "[ -e "/dev/xvdf" ] && mkfs.ext4 -L log /dev/xvdf"

[Install]
WantedBy=multi-user.target
`

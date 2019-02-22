package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
Before=docker.service var-log.mount

[Service]
Type=oneshot
ExecStart=/bin/bash -c "[ -e "/dev/xvdf" ] && /usr/sbin/mkfs.ext4 -L log /dev/xvdf"

[Install]
WantedBy=multi-user.target
`

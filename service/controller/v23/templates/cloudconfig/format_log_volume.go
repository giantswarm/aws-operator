package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
Before=docker.service var-log.mount

[Service]
Type=oneshot
ExecStart=/bin/bash -c "[ -e "/dev/xvdf" ] && /usr/sbin/mkfs.ext4 -f /dev/xvdf -L log"

[Install]
WantedBy=multi-user.target
`

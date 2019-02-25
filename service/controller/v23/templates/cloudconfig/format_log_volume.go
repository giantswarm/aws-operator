package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
Before=var-log.mount local-fs.target

[Service]
Type=oneshot
RemainAfterExit=yes

ExecStart=/bin/bash -c "[ -e "/dev/xvdf" ] && mkfs.ext4 -L log /dev/xvdf"

[Install]
WantedBy=local-fs.target
`

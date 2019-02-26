package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
Before=local-fs-pre.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=-/usr/sbin/mkfs.ext4 /dev/xvdf

[Install]
WantedBy=local-fs-pre.target
`

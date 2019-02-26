package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
DefaultDependencies=no
Before=local-fs-pre.target var-log.mount

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStartPre=-/bin/bash/ -c 'rm -rf /var/log/*'
ExecStart=-/usr/sbin/mkfs.ext4 /dev/xvdf

[Install]
WantedBy=local-fs-pre.target
`

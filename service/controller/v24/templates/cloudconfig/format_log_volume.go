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
ExecStart=-/usr/sbin/mkfs.xfs -f /dev/xvdf -L log

[Install]
WantedBy=local-fs-pre.target
`

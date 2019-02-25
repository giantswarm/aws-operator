package cloudconfig

const FormatVarLogService = `
[Unit]
Description=Formats EBS volume for log
After=dev-xvdf.device
Requires=dev-xvdf.device

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/bash -c 'if ! blkid /dev/xvdf; then /usr/sbin/mkfs.ext4 -L etcd /dev/xvdf; fi' 

[Install]
WantedBy=multi-user.target
`
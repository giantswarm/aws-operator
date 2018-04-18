package cloudconfig

const FormatEtcdVolume = `
[Unit]
Description=Formats EBS /dev/xvdh volume
Requires=dev-xvdh.device
After=dev-xvdh.device
ConditionPathExists=!/var/lib/etcd-volume-formated

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/sbin/mkfs.ext4 /dev/xvdh
ExecStartPost=/usr/bin/touch /var/lib/etcd-volume-formated

[Install]
WantedBy=multi-user.target
`

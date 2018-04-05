package cloudconfig

const FormatEtcdVolume = `
[Unit]
Description=Formats EBS /dev/xvdh volume
Requires=dev-xvdh.device
After=dev-xvdh.device

[Service]
Type=oneshot
RemainAfterExit=yes

# Do not wipe the disk if it's already being used, so the etcd data is
# persistent across reboots and updates.
Environment="LABEL=var-lib-etcd"
Environment="DEV=/dev/xvdh"
ExecStart=-/bin/bash -c "if ! findfs LABEL=$LABEL > /tmp/label.$LABEL; then wipefs -a -f $DEV && mkfs.ext4 -T news -F -L $LABEL $DEV && echo formatted file system; else echo file system already formatted; fi"

[Install]
WantedBy=multi-user.target
`

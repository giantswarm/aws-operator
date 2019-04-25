package cloudconfig

const FormatEtcdVolume = `
[Unit]
Description=Formats EBS volume for etcd
Before=docker.service var-lib-etcd.mount

[Service]
Type=oneshot
RemainAfterExit=yes

# Do not wipe the disk if it's already being used, so the etcd data is
# persistent across reboots and updates.
Environment=DEV=/dev/xvdh

# line 1: Create filesystem if does not exist.
# line 2: For compatibility with older clusters. Label existing filesystem with etcd label.
ExecStart=/bin/bash -c "\
if ! blkid $DEV; then mkfs.ext4 -L etcd $DEV; fi ;\
[ -L /dev/disk/by-label/etcd ] || e2label $DEV etcd"

[Install]
WantedBy=multi-user.target
`

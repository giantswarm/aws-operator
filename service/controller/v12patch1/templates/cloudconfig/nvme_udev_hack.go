package cloudconfig

const NVMEUdevRule = `KERNEL=="nvme[0-9]*n[0-9]*", ENV{DEVTYPE}=="disk", ATTRS{model}=="Amazon Elastic Block Store", PROGRAM="/opt/ebs-nvme-mapping /dev/%k", SYMLINK+="%c"
`

const NVMEUdevScript = `#!/bin/bash
vol=$(nvme id-ctrl --raw-binary "$1" | cut -c3073-3104 | tr -s ' ' | sed 's/ $//g')
vol=${vol#/dev/}
if [[ -n "$vol" ]]; then
    echo ${vol/xvd/sd} ${vol/sd/xvd}
fi
`

const NVMEUdevTriggerUnit = `[Unit]
Description=Reload AWS EBS NVMe rules
Requires=coreos-setup-environment.service
After=coreos-setup-environment.service
Before=user-config.target
[Service]
Type=oneshot
RemainAfterExit=yes
EnvironmentFile=-/etc/environment
ExecStart=/usr/bin/udevadm control --reload-rules
ExecStart=/usr/bin/udevadm trigger -y "nvme[0-9]*n[0-9]*"
ExecStart=/usr/bin/udevadm settle
`

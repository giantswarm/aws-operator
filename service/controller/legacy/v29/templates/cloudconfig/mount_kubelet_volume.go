package cloudconfig

const EphemeralVarLibKubeletMount = `
[Unit]
Description=kubelet volume
DefaultDependencies=no

[Mount]
What=/dev/disk/by-label/kubelet
Where=/var/lib/kubelet
Type=ext4

[Install]
WantedBy=local-fs-pre.target
`

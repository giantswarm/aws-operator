package template

const EphemeralVarLibContainerdMount = `
[Unit]
Description=Mount ephemeral volume on /var/lib/containerd
[Mount]
What=/dev/disk/by-label/containerd
Where=/var/lib/containerd
Type=xfs
[Install]
RequiredBy=local-fs.target
`

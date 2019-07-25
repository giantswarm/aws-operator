package cloudconfig

const EphemeralVarLibDockerMount = `
[Unit]
Description=Mount ephemeral volume on /var/lib/docker
[Mount]
What=/dev/disk/by-label/docker
Where=/var/lib/docker
Type=ext4
[Install]
RequiredBy=local-fs.target
`

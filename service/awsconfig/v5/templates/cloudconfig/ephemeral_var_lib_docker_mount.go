package cloudconfig

const EphemeralVarLibDockerMount = `
[Unit]
Description=Mount ephemeral volume on /var/lib/docker

[Mount]
What=/dev/xvdb
Where=/var/lib/docker
Type=xfs

[Install]
RequiredBy=local-fs.target
`

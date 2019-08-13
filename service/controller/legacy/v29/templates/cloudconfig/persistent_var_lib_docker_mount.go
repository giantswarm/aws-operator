package cloudconfig

const PersistentVarLibDockerMount = `
[Unit]
Description=Mount persistent volume on /var/lib/docker
[Mount]
What=/dev/disk/by-label/docker
Where=/var/lib/docker
Type=ext4
[Install]
RequiredBy=local-fs.target
`

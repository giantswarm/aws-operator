package cloudconfig

const PersistentVarLibDockerMountTemplate = `
[Unit]
Description=Mount persistent volume on /var/lib/docker

[Mount]
What=/dev/xvdh
Where=/var/lib/docker
Type=xfs

[Install]
RequiredBy=local-fs.target
`

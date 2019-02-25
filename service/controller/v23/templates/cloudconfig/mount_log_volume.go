package cloudconfig

const EphemeralVarLogMount = `
[Unit]
Description=log data volume

[Mount]
What=/dev/xvdf
Where=/var/log
`

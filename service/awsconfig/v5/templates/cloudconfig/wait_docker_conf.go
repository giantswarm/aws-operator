package cloudconfig

const WaitDockerConfTemplate = `
[Unit]
After=var-lib-docker.mount
Requires=var-lib-docker.mount
`

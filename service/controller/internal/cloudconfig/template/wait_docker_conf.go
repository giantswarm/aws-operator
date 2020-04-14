package template

const WaitDockerConf = `
[Unit]
After=var-lib-docker.mount
Requires=var-lib-docker.mount
`

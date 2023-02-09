package template

const PullPauseContainer = `
[Unit]
Description=Pull pause container with containerd using docker image pull secret
After=docker.service
Wants=docker.service

[Service]
Type=oneshot
ExecStart=/opt/bin/pull-image docker.io/giantswarm/pause:3.7
Restart=on-failure

[Install]
WantedBy=multi-user.target
`

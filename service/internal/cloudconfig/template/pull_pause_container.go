package template

const PullPauseContainer = `
[Unit]
Description=Pull pause container with containerd using docker image pull secret
After=containerd.service
Wants=containerd.service

[Service]
Type=oneshot
ExecStart=/bin/bash -c 'crictl pull --auth "$(cat /root/.docker/config.json | jq ".auths[\"https://index.docker.io/v1/\"].auth" -r)" docker.io/giantswarm/pause:3.7'
Restart=on-failure

[Install]
WantedBy=multi-user.target
`

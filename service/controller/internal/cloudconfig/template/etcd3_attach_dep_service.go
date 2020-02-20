package cloudconfig


const Etcd3AttachDepService = `
[Unit]
Description=Attach etcd dependencies
Requires=network.target
After=network.target

[Service]
Environment="IMAGE={{.DockerRegistry}}/giantswarm/aws-attach-etcd-dep:{{.Etcd3AttachDepDockerImage}}"
Environment="NAME=%p.service"
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/bash -c "docker run --rm -i \
      -v /dev:/dev \
      --privileged \
      --name ${NAME} \
      ${IMAGE} \
      --eni-device-index=1 \
      --eni-tag-key=Name \
      --eni-tag-value={{.Cluster.ID}}-master{{.MasterID}}-etcd \
      --volume-device-name=/dev/xvdh \
      --volume-device-filesystem-type=ext4 \
      --volume-tag-key=Name \
      --volume-tag-value={{.Cluster.ID}}-master{{.MasterID}}-etcd"
[Install]
WantedBy=multi-user.target
`


package template

const Etcd3AttachDepService = `
[Unit]
Description=Attach etcd dependencies
Requires=network.target
After=network.target
Before=k8s-kubelet.service

[Service]
# image is from https://github.com/giantswarm/aws-attach-etcd-dep
Environment="IMAGE={{ .RegistryDomain }}/giantswarm/aws-attach-etcd-dep:b49bd45ca0138e51270bb4ef726d7f646c1e5c21"
Environment="NAME=%p.service"
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/bash -c "docker run --rm -i \
      -v /dev:/dev \
      -v /etc/systemd/network:/etc/systemd/network \
      --privileged \
      --name ${NAME} \
      ${IMAGE} \
      --eni-device-index=1 \
      --eni-tag-key=Name \
      --eni-tag-value={{ .MasterENIName }} \
      --volume-device-name=/dev/xvdh \
      --volume-device-filesystem-type=ext4 \
      --volume-device-label=etcd \
      --volume-tag-key=Name \
      --volume-tag-value={{ .MasterEtcdVolumeName }}"
ExecStartPost=/usr/bin/systemctl daemon-reload
ExecStartPost=/usr/bin/systemctl restart systemd-networkd

[Install]
WantedBy=multi-user.target
`

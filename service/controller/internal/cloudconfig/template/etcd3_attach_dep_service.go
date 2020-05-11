package template

const Etcd3AttachDepService = `
[Unit]
Description=Attach etcd dependencies
Requires=network.target
After=network.target

[Service]
# image is from https://github.com/giantswarm/aws-attach-etcd-dep
Environment="IMAGE={{ .RegistryDomain }}/giantswarm/aws-attach-etcd-dep:40c13721400c3e824b908e081c12666fd81cb0b9"
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
#ExecStartPost=/usr/bin/systemctl daemon-reload
#ExecStartPost=/usr/bin/systemctl restart systemd-networkd

[Install]
WantedBy=multi-user.target
`

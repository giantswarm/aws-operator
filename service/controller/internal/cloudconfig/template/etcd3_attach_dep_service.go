package template

const Etcd3AttachDepService = `
[Unit]
Description=Attach etcd dependencies
Requires=network.target
After=network.target

[Service]
# image is from https://github.com/giantswarm/aws-attach-etcd-dep
Environment="IMAGE={{ .RegistryDomain }}/giantswarm/aws-attach-etcd-dep:b5d570e051a53c427dd86e7901cfb10b76b2e1b1"
Environment="NAME=%p.service"
Type=oneshot
RemainAfterExit=yes
ExecStart=/bin/bash -c "docker run --rm -i \
      -v /dev:/dev \
      --privileged \
      --name ${NAME} \
      ${IMAGE} \
      --eni-configure-routing=true \
      --eni-device-index=1 \
      --eni-tag-key=Name \
      --eni-tag-value={{ .MasterENIName }} \
      --volume-device-name=/dev/xvdh \
      --volume-device-filesystem-type=ext4 \
      --volume-device-label=etcd \
      --volume-tag-key=Name \
      --volume-tag-value={{ .MasterEtcdVolumeName }}"
[Install]
WantedBy=multi-user.target
`

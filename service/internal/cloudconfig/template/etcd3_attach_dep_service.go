package template

const Etcd3AttachDepService = `
[Unit]
Description=Attach etcd dependencies
Requires=network.target
After=network.target
Before=k8s-kubelet.service

[Service]
# image is from https://github.com/giantswarm/aws-attach-etcd-dep
Environment="IMAGE={{ .RegistryDomain }}/giantswarm/aws-attach-etcd-dep:0.1.0"
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
# use 'networkctl reload' instead of restarting systemd-networkd to avoid bug in systemd 
# https://github.com/systemd/systemd/issues/18108
ExecStartPost=/usr/bin/networkctl reload

[Install]
WantedBy=multi-user.target
`

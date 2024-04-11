package template

const Etcd3AttachDepService = `
[Unit]
Description=Attach etcd dependencies
Requires=network.target
After=network.target
Before=k8s-kubelet.service

[Service]
# image is from https://github.com/giantswarm/aws-attach-etcd-dep
Environment="IMAGE=quay.io/giantswarm/aws-attach-etcd-dep:0.4.0-cb5b8660cad92753bf6d83d9ca5fb46cc233ae44"
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

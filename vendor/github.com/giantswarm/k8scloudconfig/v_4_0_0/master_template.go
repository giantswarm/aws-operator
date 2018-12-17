package v_4_0_0

const MasterTemplate = `---
ignition:
  version: "2.2.0"
passwd:
  users:
{{ range $index, $user := .Cluster.Kubernetes.SSH.UserList }}
    - name: {{ $user.Name }}
      shell: "/bin/bash"
      groups:
        - "sudo"
        - "docker"
{{ if ne $user.PublicKey "" }}
      sshAuthorizedKeys:
        - "{{ $user.PublicKey }}"
{{ end }}
{{ end }}

systemd:
  units:
  - name: audit-rules.service
    enabled: true
    dropins:
    - name: 10-Wait-For-Docker.conf
      contents: |
        [Service]
        ExecStartPre=/bin/bash -c "while [ ! -f /etc/audit/rules.d/10-docker.rules ]; do echo 'Waiting for /etc/audit/rules.d/10-docker.rules to be written' && sleep 1; done"
  {{range .Extension.Units}}
  - name: {{.Metadata.Name}}
    enabled: {{.Metadata.Enabled}}
    contents: |
      {{range .Content}}{{.}}
      {{end}}{{end}}
  - name: wait-for-domains.service
    enabled: true
    contents: |
      [Unit]
      Description=Wait for etcd and k8s API domains to be available
      [Service]
      Type=oneshot
      ExecStart=/opt/wait-for-domains
      [Install]
      WantedBy=multi-user.target
  - name: os-hardeing.service
    enabled: true
    contents: |
      [Unit]
      Description=Apply os hardening
      [Service]
      Type=oneshot
      ExecStartPre=-/bin/bash -c "gpasswd -d core rkt; gpasswd -d core docker; gpasswd -d core wheel"
      ExecStartPre=/bin/bash -c "until [ -f '/etc/sysctl.d/hardening.conf' ]; do echo Waiting for sysctl file; sleep 1s;done;"
      ExecStart=/usr/sbin/sysctl -p /etc/sysctl.d/hardening.conf
      [Install]
      WantedBy=multi-user.target
  - name: k8s-setup-kubelet-config.service
    enabled: true
    contents: |
      [Unit]
      Description=k8s-setup-kubelet-config Service
      After=k8s-setup-network-env.service docker.service
      Requires=k8s-setup-network-env.service docker.service
      [Service]
      Type=oneshot
      RemainAfterExit=yes
      TimeoutStartSec=0
      EnvironmentFile=/etc/network-environment
      ExecStart=/bin/bash -c '/usr/bin/envsubst </etc/kubernetes/config/kubelet.yaml.tmpl >/etc/kubernetes/config/kubelet.yaml'
      [Install]
      WantedBy=multi-user.target
  - name: docker.service
    enabled: true
    contents: |
    dropins:
      - name: 10-giantswarm-extra-args.conf
        contents: |
          [Service]
          Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=cgroupfs --log-opt max-size=25m --log-opt max-file=2 --log-opt labels=io.kubernetes.container.hash,io.kubernetes.container.name,io.kubernetes.pod.name,io.kubernetes.pod.namespace,io.kubernetes.pod.uid"
          Environment="DOCKER_OPT_BIP=--bip={{.Cluster.Docker.Daemon.CIDR}}"
          Environment="DOCKER_OPTS=--live-restore --icc=false --userland-proxy=false"
  - name: k8s-setup-network-env.service
    enabled: true
    contents: |
      [Unit]
      Description=k8s-setup-network-env Service
      Wants=network.target docker.service
      After=network.target docker.service
      [Service]
      Type=oneshot
      RemainAfterExit=yes
      TimeoutStartSec=0
      Environment="IMAGE={{.Cluster.Kubernetes.NetworkSetup.Docker.Image}}"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/mkdir -p /opt/bin/
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host -v /etc:/etc --name $NAME $IMAGE
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
      [Install]
      WantedBy=multi-user.target
  - name: etcd3.service
    enabled: true
    contents: |
      [Unit]
      Description=etcd3
      Requires=k8s-setup-network-env.service
      After=k8s-setup-network-env.service
      Conflicts=etcd.service etcd2.service
      StartLimitIntervalSec=0
      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      LimitNOFILE=40000
      Environment=IMAGE={{ .RegistryDomain }}/{{ .Images.Etcd }}
      Environment=NAME=%p.service
      EnvironmentFile=/etc/network-environment
      ExecStartPre=-/usr/bin/docker stop  $NAME
      ExecStartPre=-/usr/bin/docker rm  $NAME
      ExecStartPre=-/usr/bin/docker pull $IMAGE
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/etcd/server-ca.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/etcd/server-ca.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/etcd/server-crt.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/etcd/server-crt.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/etcd/server-key.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/etcd/server-key.pem to be written' && sleep 1; done"
      ExecStart=/usr/bin/docker run \
          -v /etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt \
          -v /etc/kubernetes/ssl/etcd/:/etc/etcd \
          -v /var/lib/etcd/:/var/lib/etcd  \
          --net=host  \
          --name $NAME \
          $IMAGE \
          etcd \
          --name etcd0 \
          --trusted-ca-file /etc/etcd/server-ca.pem \
          --cert-file /etc/etcd/server-crt.pem \
          --key-file /etc/etcd/server-key.pem\
          --client-cert-auth=true \
          --peer-trusted-ca-file /etc/etcd/server-ca.pem \
          --peer-cert-file /etc/etcd/server-crt.pem \
          --peer-key-file /etc/etcd/server-key.pem \
          --peer-client-cert-auth=true \
          --advertise-client-urls=https://{{ .Cluster.Etcd.Domain }}:{{ .EtcdPort }} \
          --initial-advertise-peer-urls=https://127.0.0.1:2380 \
          --listen-client-urls=https://0.0.0.0:2379 \
          --listen-peer-urls=https://${DEFAULT_IPV4}:2380 \
          --initial-cluster-token k8s-etcd-cluster \
          --initial-cluster etcd0=https://127.0.0.1:2380 \
          --initial-cluster-state new \
          --data-dir=/var/lib/etcd \
          --enable-v2
      [Install]
      WantedBy=multi-user.target
  - name: etcd3-defragmentation.service
    enabled: false
    contents: |
      [Unit]
      Description=etcd defragmentation job
      After=docker.service etcd3.service
      Requires=docker.service etcd3.service
      [Service]
      Type=oneshot
      EnvironmentFile=/etc/network-environment
      Environment=IMAGE={{ .RegistryDomain }}/{{ .Images.Etcd }}
      Environment=NAME=%p.service
      ExecStartPre=-/usr/bin/docker stop  $NAME
      ExecStartPre=-/usr/bin/docker rm  $NAME
      ExecStartPre=-/usr/bin/docker pull $IMAGE
      ExecStart=/usr/bin/docker run \
        -v /etc/kubernetes/ssl/etcd/:/etc/etcd \
        --net=host  \
        -e ETCDCTL_API=3 \
        --name $NAME \
        $IMAGE \
        etcdctl \
        --endpoints https://127.0.0.1:2379 \
        --cacert /etc/etcd/server-ca.pem \
        --cert /etc/etcd/server-crt.pem \
        --key /etc/etcd/server-key.pem \
        defrag
      [Install]
      WantedBy=multi-user.target
  - name: etcd3-defragmentation.timer
    enabled: true
    contents: |
      [Unit]
      Description=Execute etcd3-defragmentation every day at 3.30AM UTC
      [Timer]
      OnCalendar=*-*-* 03:30:00 UTC
      [Install]
      WantedBy=multi-user.target
  - name: k8s-kubelet.service
    enabled: true
    contents: |
      [Unit]
      Wants=k8s-setup-network-env.service k8s-setup-kubelet-config.service
      After=k8s-setup-network-env.service k8s-setup-kubelet-config.service
      Description=k8s-kubelet
      StartLimitIntervalSec=0
      [Service]
      TimeoutStartSec=300
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE={{ .RegistryDomain }}/{{ .Images.Kubernetes }}"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/bin/sh -c "/usr/bin/docker run --rm --pid=host --net=host --privileged=true \
      {{ range .Hyperkube.Kubelet.Docker.RunExtraArgs -}}
      {{ . }} \
      {{ end -}}
      -v /:/rootfs:ro,rshared \
      -v /sys:/sys:ro \
      -v /dev:/dev:rw \
      -v /var/log:/var/log:rw \
      -v /run/calico/:/run/calico/:rw \
      -v /run/docker/:/run/docker/:rw \
      -v /run/docker.sock:/run/docker.sock:rw \
      -v /usr/lib/os-release:/etc/os-release \
      -v /usr/share/ca-certificates/:/etc/ssl/certs \
      -v /var/lib/calico/:/var/lib/calico \
      -v /var/lib/docker/:/var/lib/docker:rw,rshared \
      -v /var/lib/kubelet/:/var/lib/kubelet:rw,rshared \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      -v /etc/kubernetes/kubeconfig/:/etc/kubernetes/kubeconfig/ \
      -v /etc/kubernetes/manifests/:/etc/kubernetes/manifests/ \
      -v /etc/cni/net.d/:/etc/cni/net.d/ \
      -v /opt/cni/bin/:/opt/cni/bin/ \
      -v /usr/sbin/iscsiadm:/usr/sbin/iscsiadm \
      -v /etc/iscsi/:/etc/iscsi/ \
      -v /dev/disk/by-path/:/dev/disk/by-path/ \
      -v /dev/mapper/:/dev/mapper/ \
      -v /lib/modules:/lib/modules \
      -v /usr/sbin/mkfs.xfs:/usr/sbin/mkfs.xfs \
      -v /usr/lib64/libxfs.so.0:/usr/lib/libxfs.so.0 \
      -v /usr/lib64/libxcmd.so.0:/usr/lib/libxcmd.so.0 \
      -v /usr/lib64/libreadline.so.7:/usr/lib/libreadline.so.7 \
      -e ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/etcd/server-ca.pem \
      -e ETCD_CERT_FILE=/etc/kubernetes/ssl/etcd/server-crt.pem \
      -e ETCD_KEY_FILE=/etc/kubernetes/ssl/etcd/server-key.pem \
      --name $NAME \
      $IMAGE \
      /hyperkube kubelet \
      {{ range .Hyperkube.Kubelet.Docker.CommandExtraArgs -}}
      {{ . }} \
      {{ end -}}
      --node-ip=${DEFAULT_IPV4} \
      --config=/etc/kubernetes/config/kubelet.yaml \
      --containerized \
      --enable-server \
      --logtostderr=true \
      --cloud-provider={{.Cluster.Kubernetes.CloudProvider}} \
      --network-plugin=cni \
      --register-node=true \
      --register-with-taints=node-role.kubernetes.io/master=:NoSchedule \
      --kubeconfig=/etc/kubernetes/kubeconfig/kubelet.yaml \
      --node-labels="node-role.kubernetes.io/master,role=master,ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
      --v=2"
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
      [Install]
      WantedBy=multi-user.target
      Requires=network-online.target
      After=network-online.target
  - name: etcd2.service
    enabled: false
    mask: true
  - name: update-engine.service
    enabled: false
    mask: true
  - name: locksmithd.service
    enabled: false
    mask: true
  - name: fleet.service
    enabled: false
    mask: true
  - name: fleet.socket
    enabled: false
    mask: true
  - name: flanneld.service
    enabled: false
    mask: true
  - name: systemd-networkd-wait-online.service
    enabled: false
    mask: true
  - name: k8s-addons.service
    enabled: true
    contents: |
      [Unit]
      Description=Kubernetes Addons
      Wants=k8s-kubelet.service
      After=k8s-kubelet.service
      [Service]
      Type=oneshot
      EnvironmentFile=/etc/network-environment
      ExecStart=/opt/k8s-addons
      # https://github.com/kubernetes/kubernetes/issues/71078
      ExecStartPost=/usr/bin/systemctl restart k8s-kubelet.service
      [Install]
      WantedBy=multi-user.target

storage:
  files:
    - path: /etc/ssh/trusted-user-ca-keys.pem
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;base64,{{ index .Files "conf/trusted-user-ca-keys.pem" }}"

    {{- if not .DisableCalico }}
    - path: /srv/calico-all.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/calico-all.yaml" }}"
    {{- end }}

    {{- if not .DisableCoreDNS }}
    - path: /srv/coredns.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/coredns.yaml" }}"
    {{- end }}

    {{- if not .DisableIngressController }}
    - path: /srv/default-backend-dep.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/default-backend-dep.yaml" }}"

    - path: /srv/default-backend-svc.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/default-backend-svc.yaml" }}"

    - path: /srv/ingress-controller-cm.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/ingress-controller-cm.yaml" }}"

    - path: /srv/ingress-controller-dep.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/ingress-controller-dep.yaml" }}"
    {{- end }}

    {{- if not .DisableIngressControllerService }}
    - path: /srv/ingress-controller-svc.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/ingress-controller-svc.yaml" }}"
    {{- end }}

    - path: /etc/kubernetes/config/proxy-config.yml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "config/kube-proxy.yaml" }}"

    - path: /srv/kube-proxy-config.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "config/kube-proxy.yaml" }}"

    - path: /srv/kube-proxy-sa.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/kube-proxy-sa.yaml" }}"

    - path: /srv/kube-proxy-ds.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/kube-proxy-ds.yaml" }}"

    - path: /srv/rbac_bindings.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/rbac_bindings.yaml" }}"

    - path: /srv/rbac_roles.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/rbac_roles.yaml" }}"

    - path: /srv/priority_classes.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/priority_classes.yaml" }}"

    - path: /srv/psp_policies.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/psp_policies.yaml" }}"

    - path: /srv/psp_roles.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/psp_roles.yaml" }}"

    - path: /srv/psp_binding.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/psp_bindings.yaml" }}"

    - path: /opt/wait-for-domains
      filesystem: root
      mode: 0544
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "conf/wait-for-domains" }}"

    - path: /opt/k8s-addons
      filesystem: root
      mode: 0544
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "conf/k8s-addons" }}"

    - path: /etc/kubernetes/kubeconfig/addons.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "kubeconfig/addons.yaml" }}"

    - path: /etc/kubernetes/config/proxy-kubeconfig.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "kubeconfig/kube-proxy-master.yaml" }}"

    - path: /etc/kubernetes/kubeconfig/kube-proxy.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "kubeconfig/kube-proxy-master.yaml" }}"

    - path: /etc/kubernetes/config/kubelet.yaml.tmpl
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "config/kubelet.yaml.tmpl" }}"

    - path: /etc/kubernetes/kubeconfig/kubelet.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "kubeconfig/kubelet-master.yaml" }}"

    - path: /etc/kubernetes/kubeconfig/controller-manager.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "kubeconfig/controller-manager.yaml" }}"

    - path: /etc/kubernetes/config/scheduler.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "config/scheduler.yaml" }}"

    - path: /etc/kubernetes/kubeconfig/scheduler.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "kubeconfig/scheduler.yaml" }}"

    {{ if not .DisableEncryptionAtREST -}}
    - path: /etc/kubernetes/encryption/k8s-encryption-config.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "k8s-resource/k8s-encryption-config.yaml" }}"

    {{ end -}}
    - path: /etc/kubernetes/policies/audit-policy.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "policies/audit-policy.yaml" }}"

    - path: /etc/kubernetes/manifests/k8s-api-server.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "manifests/k8s-api-server.yaml" }}"

    - path: /etc/kubernetes/manifests/k8s-controller-manager.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "manifests/k8s-controller-manager.yaml" }}"

    - path: /etc/kubernetes/manifests/k8s-scheduler.yaml
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "manifests/k8s-scheduler.yaml" }}"

    - path: /etc/ssh/sshd_config
      filesystem: root
      mode: 0644
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "conf/sshd_config" }}"

    - path: /etc/sysctl.d/hardening.conf
      filesystem: root
      mode: 0600
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "conf/hardening.conf" }}"

    - path: /etc/audit/rules.d/10-docker.rules
      filesystem: root
      mode: 0600
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "conf/10-docker.rules" }}"

    - path: /etc/modules-load.d/ip_vs.conf
      filesystem: root
      mode: 0600
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{  index .Files "conf/ip_vs.conf" }}"

    {{ range .Extension.Files -}}
    - path: {{ .Metadata.Path }}
      filesystem: root
      user:
        name: {{ .Metadata.Owner.User }}
      group:
        name: {{ .Metadata.Owner.Group }}
      mode: {{printf "%#o" .Metadata.Permissions}}
      contents:
        source: "data:text/plain;charset=utf-8;base64,{{ .Content }}"
        {{ if .Metadata.Compression }}
        compression: gzip
        {{end}}
    {{ end -}}

{{ range .Extension.VerbatimSections }}
{{ .Content }}
{{ end }}
`

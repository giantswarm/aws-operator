package v_3_5_3

const WorkerTemplate = `#cloud-config
users:
{{ range $index, $user := .Cluster.Kubernetes.SSH.UserList }}  - name: {{ $user.Name }}
    groups:
      - "sudo"
      - "docker"
{{ if ne $user.PublicKey "" }}
    ssh-authorized-keys:
       - "{{ $user.PublicKey }}"
{{ end }}
{{end}}
write_files:
- path: /etc/ssh/trusted-user-ca-keys.pem
  owner: root
  permissions: 644
  content: |
    {{ .SSOPublicKey }}
- path: /etc/kubernetes/config/proxy-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: proxy
      user:
        client-certificate: /etc/kubernetes/ssl/worker-crt.pem
        client-key: /etc/kubernetes/ssl/worker-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/worker-ca.pem
        server: https://{{.Cluster.Kubernetes.API.Domain}}
    contexts:
    - context:
        cluster: local
        user: proxy
      name: service-account-context
    current-context: service-account-context
- path: /etc/kubernetes/config/kubelet-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: kubelet
      user:
        client-certificate: /etc/kubernetes/ssl/worker-crt.pem
        client-key: /etc/kubernetes/ssl/worker-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/worker-ca.pem
        server: https://{{.Cluster.Kubernetes.API.Domain}}
    contexts:
    - context:
        cluster: local
        user: kubelet
      name: service-account-context
    current-context: service-account-context
- path: /opt/wait-for-domains
  permissions: 0544
  content: |
      #!/bin/bash
      domains="{{.Cluster.Etcd.Domain}} {{.Cluster.Kubernetes.API.Domain}}"

      for domain in $domains; do
        until nslookup $domain; do
            echo "Waiting for domain $domain to be available"
            sleep 5
        done

        echo "Successfully resolved domain $domain"
      done

- path: /etc/ssh/sshd_config
  owner: root
  permissions: 0600
  content: |
    # Use most defaults for sshd configuration.
    UsePrivilegeSeparation sandbox
    Subsystem sftp internal-sftp
    ClientAliveInterval 180
    UseDNS no
    UsePAM yes
    PrintLastLog no # handled by PAM
    PrintMotd no # handled by PAM
    # Non defaults (#100)
    ClientAliveCountMax 2
    PasswordAuthentication no
    TrustedUserCAKeys /etc/ssh/trusted-user-ca-keys.pem
- path: /etc/sysctl.d/hardening.conf
  owner: root
  permissions: 0600
  content: |
    kernel.kptr_restrict = 2
    kernel.sysrq = 0
    net.ipv4.conf.all.log_martians = 1
    net.ipv4.conf.all.send_redirects = 0
    net.ipv4.conf.default.accept_redirects = 0
    net.ipv4.conf.default.log_martians = 1
    net.ipv4.tcp_timestamps = 0
    net.ipv6.conf.all.accept_redirects = 0
    net.ipv6.conf.default.accept_redirects = 0

- path: /etc/audit/rules.d/10-docker.rules
  owner: root
  permissions: 644
  content: |
    -w /usr/bin/docker -k docker
    -w /var/lib/docker -k docker
    -w /etc/docker -k docker
    -w /etc/systemd/system/docker.service.d/10-giantswarm-extra-args.conf -k docker
    -w /etc/systemd/system/docker.service.d/01-wait-docker.conf -k docker
    -w /usr/lib/systemd/system/docker.service -k docker
    -w /usr/lib/systemd/system/docker.socket -k docker

- path: /etc/systemd/system/audit-rules.service.d/10-Wait-For-Docker.conf
  owner: root
  permissions: 644
  content: |
    [Service]
    ExecStartPre=/bin/bash -c "while [ ! -f /etc/audit/rules.d/10-docker.rules ]; do echo 'Waiting for /etc/audit/rules.d/10-docker.rules to be written' && sleep 1; done"

{{range .Extension.Files}}
- path: {{.Metadata.Path}}
  owner: {{.Metadata.Owner}}
  {{ if .Metadata.Encoding }}
  encoding: {{.Metadata.Encoding}}
  {{ end }}
  permissions: {{printf "%#o" .Metadata.Permissions}}
  content: |
    {{range .Content}}{{.}}
    {{end}}{{end}}

- path: /etc/modules-load.d/ip_vs.conf
  owner: root
  permissions: 644
  content: |
    ip_vs
    ip_vs_rr
    ip_vs_wrr
    ip_vs_sh
    nf_conntrack_ipv4

coreos:
  units:
  {{range .Extension.Units}}
  - name: {{.Metadata.Name}}
    enable: {{.Metadata.Enable}}
    command: {{.Metadata.Command}}
    content: |
      {{range .Content}}{{.}}
      {{end}}{{end}}
  - name: wait-for-domains.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Wait for etcd and k8s API domains to be available

      [Service]
      Type=oneshot
      ExecStart=/opt/wait-for-domains

      [Install]
      WantedBy=multi-user.target
  - name: os-hardeing.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Apply os hardening

      [Service]
      Type=oneshot
      ExecStartPre=-/bin/bash -c "gpasswd -d core rkt; gpasswd -d core docker; gpasswd -d core wheel"
      ExecStartPre=/bin/bash -c "until [ -f '/etc/sysctl.d/hardening.conf' ]; do echo Waiting for sysctl file; sleep 1s;done;"
      ExecStart=/usr/sbin/sysctl -p /etc/sysctl.d/hardening.conf

      [Install]
      WantedBy=multi-user.target
  - name: update-engine.service
    enable: false
    command: stop
    mask: true
  - name: locksmithd.service
    enable: false
    command: stop
    mask: true
  - name: etcd2.service
    enable: false
    command: stop
    mask: true
  - name: fleet.service
    enable: false
    command: stop
    mask: true
  - name: fleet.socket
    enable: false
    command: stop
    mask: true
  - name: flanneld.service
    enable: false
    command: stop
    mask: true
  - name: systemd-networkd-wait-online.service
    mask: true
  - name: docker.service
    enable: true
    command: start
    drop-ins:
    - name: 10-giantswarm-extra-args.conf
      content: |
        [Service]
        Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=cgroupfs --log-opt max-size=25m --log-opt max-file=2 --log-opt labels=io.kubernetes.container.hash,io.kubernetes.container.name,io.kubernetes.pod.name,io.kubernetes.pod.namespace,io.kubernetes.pod.uid"
        Environment="DOCKER_OPT_BIP=--bip={{.Cluster.Docker.Daemon.CIDR}}"
        Environment="DOCKER_OPTS=--live-restore --icc=false --userland-proxy=false"
  - name: k8s-setup-network-env.service
    enable: true
    command: start
    content: |
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
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host -v /etc:/etc --name $NAME $IMAGE
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-kubelet.service
    enable: true
    command: start
    content: |
      [Unit]
      Wants=k8s-setup-network-env.service
      After=k8s-setup-network-env.service
      Description=k8s-kubelet
      StartLimitIntervalSec=0

      [Service]
      TimeoutStartSec=300
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE={{ .RegistryDomain }}/giantswarm/hyperkube:v1.10.11"
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
      -v /var/lib/docker/:/var/lib/docker:rw,rshared \
      -v /var/lib/kubelet/:/var/lib/kubelet:rw,rshared \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
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
      -e ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/etcd/client-ca.pem \
      -e ETCD_CERT_FILE=/etc/kubernetes/ssl/etcd/client-crt.pem \
      -e ETCD_KEY_FILE=/etc/kubernetes/ssl/etcd/client-key.pem \
      --name $NAME \
      $IMAGE \
      /hyperkube kubelet \
      {{ range .Hyperkube.Kubelet.Docker.CommandExtraArgs -}}
      {{ . }} \
      {{ end -}}
      --address=${DEFAULT_IPV4} \
      --port={{.Cluster.Kubernetes.Kubelet.Port}} \
      --node-ip=${DEFAULT_IPV4} \
      --containerized \
      --enable-server \
      --logtostderr=true \
      --machine-id-file=/rootfs/etc/machine-id \
      --cadvisor-port=4194 \
      --cloud-provider={{.Cluster.Kubernetes.CloudProvider}} \
      --healthz-bind-address=${DEFAULT_IPV4} \
      --healthz-port=10248 \
      --cluster-dns={{.Cluster.Kubernetes.DNS.IP}} \
      --cluster-domain={{.Cluster.Kubernetes.Domain}} \
      --network-plugin=cni \
      --register-node=true \
      --allow-privileged=true \
      --feature-gates=ExpandPersistentVolumes=true,PodPriority=true,CustomResourceSubresources=true \
      --kubeconfig=/etc/kubernetes/config/kubelet-kubeconfig.yml \
      --node-labels="ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
      --kube-reserved="cpu=200m,memory=250Mi" \
      --system-reserved="cpu=150m,memory=250Mi" \
      --eviction-soft='memory.available<500Mi' \
      --eviction-hard='memory.available<350Mi' \
      --eviction-soft-grace-period='memory.available=5s' \
      --eviction-max-pod-grace-period=60 \
      --enforce-node-allocatable=pods \
      --v=2"
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME

  update:
    reboot-strategy: off

{{ range .Extension.VerbatimSections }}
{{ .Content }}
{{ end }}
`

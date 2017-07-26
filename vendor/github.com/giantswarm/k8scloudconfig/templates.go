package cloudconfig

const (
	MasterTemplate = `#ignition
ignition:
  version: 2.0.0
systemd:
  units:
    {{range .Extension.Units}}
    - name: {{.Metadata.Name}}
      enable: {{.Metadata.Enable}}
      contents: |
        {{range .Content}}{{.}}
        {{end}}{{end}}
    - name: kubelet.service
      enable: true
      contents: |
        [Service]
        Restart=always
        RestartSec=0
        TimeoutStopSec=10
        EnvironmentFile=/etc/network-environment
        Environment="IMAGE={{.Cluster.Kubernetes.Hyperkube.Docker.Image}}"
        Environment="NAME=%p.service"
        Environment="NETWORK_CONFIG_CONTAINER="
        ExecStartPre=/usr/bin/docker pull $IMAGE
        ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
        ExecStartPre=-/usr/bin/docker rm -f $NAME
        ExecStart=/bin/sh -c "/usr/bin/docker run --rm --pid=host --net=host --privileged=true \
        -v /:/rootfs:ro \
        -v /sys:/sys:ro \
        -v /dev:/dev:rw \
        -v /var/log/pods:/var/log/pods:rw \
        -v /run/calico/:/run/calico/:rw \
        -v /run/docker/:/run/docker/:rw \
        -v /run/docker.sock:/run/docker.sock:rw \
        -v /usr/lib/os-release:/etc/os-release \
        -v /usr/share/ca-certificates/:/etc/ssl/certs \
        -v /var/lib/docker/:/var/lib/docker:rw \
        -v /var/lib/kubelet/:/var/lib/kubelet:rw,rslave \
        -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
        -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
        -v /etc/cni/net.d/:/etc/cni/net.d/ \
        -v /opt/cni/bin/:/opt/cni/bin/ \
        -e ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/etcd/server-ca.pem \
        -e ETCD_CERT_FILE=/etc/kubernetes/ssl/etcd/server-crt.pem \
        -e ETCD_KEY_FILE=/etc/kubernetes/ssl/etcd/server-key.pem \
        --name $NAME \
        $IMAGE \
        /hyperkube kubelet \
        --address=${DEFAULT_IPV4} \
        --port={{.Cluster.Kubernetes.Kubelet.Port}} \
        --hostname-override=${DEFAULT_IPV4} \
        --node-ip=${DEFAULT_IPV4} \
        --api-servers=https://{{.Cluster.Kubernetes.API.Domain}} \
        --containerized \
        --enable-server \
        --logtostderr=true \
        --machine-id-file=/rootfs/etc/machine-id \
        --cadvisor-port=4194 \
        --healthz-bind-address=${DEFAULT_IPV4} \
        --healthz-port=10248 \
        --cluster-dns={{.Cluster.Kubernetes.DNS.IP}} \
        --cluster-domain={{.Cluster.Kubernetes.Domain}} \
        --network-plugin=cni \
        --register-node=true \
        --register-schedulable=false \
        --allow-privileged=true \
        --kubeconfig=/etc/kubernetes/config/kubelet-kubeconfig.yml \
        --node-labels="role=master,kubernetes.io/hostname=${HOSTNAME},ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
        --v=2"
        ExecStop=-/usr/bin/docker stop -t 10 $NAME
        ExecStopPost=-/usr/bin/docker rm -f $NAME
        [Install]
        WantedBy=multi-user.target
    - name: kubeadm-init.service
      enable: true
      contents: |
        [Unit]
        Description=Initialize master with kubeadm
        Wants=kubelet.service
        [Service]
        Type=oneshot
	Environment=IMAGE=maru/kubeadm
        Environment=NAME=%p.service
        EnvironmentFile=/etc/network-environment
        ExecStartPre=-/usr/bin/docker stop $NAME
        ExecStartPre=-/usr/bin/docker rm $NAME
        ExecStartPre=-/usr/bin/docker pull $IMAGE
        ExecStart=/usr/bin/docker run \
            --name $NAME \
            $IMAGE \
            kubeadm init --token {{.KubeadmToken}}
        [Install]
        WantedBy=multi-user.target
    - name: wait-for-k8s-nodes.service
      enable: true
      contents: |
        [Unit]
        Description=Wait for nodes availability
        Wants=kubeadm-init.service
        [Service]
        Type=oneshot
        Environment=IMAGE={{.Cluster.Kubernetes.Kubectl.Docker.Image}}
	Environment=NAME=%p.service
	EnvironmentFile=/etc/network-environment
	ExecStartPre=-/usr/bin/docker stop $NAME
        ExecStartPre=-/usr/bin/docker pull $NAME
        ExecStart=/usr/bin/docker run \
            --name $NAME \
            $IMAGE \
            kubectl get nodes
        [Install]
        WantedBy=multi-user.target
    - name: calico-setup.service
      enable: true
      contents: |
        [Unit]
        Description=Setup self-hosted Calico
        Wants=wait-for-k8s-nodes.service
        [Service]
        Type=oneshot
	EnvironmentFile=/etc/network-environment
        ExecStart=/opt/calico-setup
    - name: k8s-addons.service
      enable: true
      contents: |
        [Unit]
        Description=Kubernetes Addons
        Wants=k8s-api-server.service
        After=k8s-api-server.service
        [Service]
        Type=oneshot
        EnvironmentFile=/etc/network-environment
        ExecStart=/opt/k8s-addons
        [Install]
        WantedBy=multi-user.target
    - name: update-engine.service
      enable: false
    - name: locksmithd.service
      mask: true
    - name: fleet.service
      mask: true
    - name: systemd-networkd-wait-online.service
      enable: true
    - name: docker.service
      enable: true
      dropins:
        - name: 10-giantswarm-extra-args.conf
          contents: |
            [Service]
            Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=cgroupfs {{.Cluster.Docker.Daemon.ExtraArgs}}"
storage:
  files:
    {{range.Extension.Files}}
    - filesystem: root
      path: {{.Metadata.Path}}
      {{if .Metadata.Encoding}}
      encoding: {{.Metadata.Encoding}}
      {{end}}
      mode: {{printf "%#o" .Metadata.Permissions}}
      contents:
        inline: |
          {{range .Content}}{{.}}
	  {{end}}{{end}}
    - filesystem: root
      path: /etc/kubernetes/config/kubelet-kubeconfig.yml
      mode: 0664
      contents:
        inline: |
          apiVersion: v1
          kind: Config
          users:
          - name: kubelet
            user:
              client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
              client-key: /etc/kubernetes/ssl/apiserver-key.pem
          clusters:
          - name: local
            cluster:
              certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
          contexts:
          - context:
              cluster: local
              user: kubelet
            name: service-account-context
          current-context: service-account-context
    - filesystem: root
      path: /srv/ingress-controller-cm.yml
      mode: 0644
      contents:
        inline: |
          kind: ConfigMap
          apiVersion: v1
          metadata:
            name: ingress-nginx
            namespace: kube-system
            labels:
              k8s-addon: ingress-nginx.addons.k8s.io
          data:
            server-name-hash-bucket-size: "1024"
            server-name-hash-max-size: "1024"
    - filesystem: root
      path: /srv/ingress-controller-dep.yml
      mode: 0644
      contents:
        inline: |
          apiVersion: extensions/v1beta1
          kind: Deployment
          metadata:
            name: nginx-ingress-controller
            namespace: kube-system
            labels:
              k8s-app: nginx-ingress-controller
            annotations:
              prometheus.io/port: '10254'
              prometheus.io/scrape: 'true'
          spec:
            replicas: 3
            strategy:
              type: RollingUpdate
              rollingUpdate:
                maxUnavailable: 2
            template:
              metadata:
                labels:
                  k8s-app: nginx-ingress-controller
                annotations:
                  scheduler.alpha.kubernetes.io/affinity: >
                    {
                      "podAntiAffinity": {
                        "preferredDuringSchedulingIgnoredDuringExecution": [
                          {
                            "labelSelector": {
                              "matchExpressions": [
                                {
                                  "key": "k8s-app",
                                  "operator": "In",
                                  "values": ["nginx-ingress-controller"]
                                }
                              ]
                            },
                            "topologyKey": "kubernetes.io/hostname",
                            "weight": 100
                          }
                        ]
                      }
                    }
              spec:
                containers:
                - name: nginx-ingress-controller
                  image: gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.7
                  args:
                  - /nginx-ingress-controller
                  - --default-backend-service=$(POD_NAMESPACE)/default-http-backend
                  - --configmap=$(POD_NAMESPACE)/ingress-nginx
                  env:
                    - name: POD_NAME
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.name
                    - name: POD_NAMESPACE
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.namespace
                  readinessProbe:
                    httpGet:
                      path: /healthz
                      port: 10254
                      scheme: HTTP
                  livenessProbe:
                    httpGet:
                      path: /healthz
                      port: 10254
                      scheme: HTTP
                    initialDelaySeconds: 10
                    timeoutSeconds: 1
                  ports:
                  - containerPort: 80
                    hostPort: 80
                  - containerPort: 443
                    hostPort: 443
    - filesystem: root
      path: /srv/ingress-controller-svc.yml
      owner: root
      permissions: 0644
      content: |
        apiVersion: v1
        kind: Service
        metadata:
          name: nginx-ingress-controller
          namespace: kube-system
          labels:
            k8s-app: nginx-ingress-controller
        spec:
          type: NodePort
          ports:
          - name: http
            port: 80
            nodePort: 30010
            protocol: TCP
            targetPort: 80
          - name: https
            port: 443
            nodePort: 30011
            protocol: TCP
            targetPort: 443
          selector:
            k8s-app: nginx-ingress-controller
    - filesystem: root
      path: /opt/calico-setup
      mode: 0644
      contents:
        inline: |
          #!/bin/bash
          KUBECTL={{.Cluster.Kubernetes.Kubectl.Docker.Image}}

          docker run -ti --rm -v /tmp:/git alpine/git clone git://github.com/projectcalico/calico.git
          docker --net=host --rm -v /srv:/srv $KUBECTL apply -f /tmp/calico/v2.3/getting-started/kubernetes/installation/hosted/kubeadm/1.6/calico.yaml
	  rm -rf /tmp/calico
    - filesystem: root
      path: /opt/k8s-addons
      mode: 0544
      contents:
        inline: |
          #!/bin/bash
          KUBECTL={{.Cluster.Kubernetes.Kubectl.Docker.Image}}

          /usr/bin/docker pull $KUBECTL

          # wait for healthy master
          while [ "$(/usr/bin/docker run --net=host --rm $KUBECTL get cs | grep Healthy | wc -l)" -ne "3" ]; do sleep 1 && echo 'Waiting for healthy k8s'; done

          # apply default storage class
          if [ -f /srv/default-storage-class.yaml ]; then
              while
                  /usr/bin/docker run --net=host --rm -v /srv:/srv $KUBECTL apply -f /srv/default-storage-class.yaml
                  [ "$?" -ne "0" ]
              do
                  echo "failed to apply /srv/default-storage-class.yaml, retrying in 5 sec"
                  sleep 5s
              done
          else
              echo "no default storage class to apply"
          fi

          # apply k8s addons
          MANIFESTS="ingress-controller-cm.yml ingress-controller-dep.yml ingress-controller-svc.yml"
          for manifest in $MANIFESTS
          do
              while
                  /usr/bin/docker run --net=host --rm -v /srv:/srv $KUBECTL apply -f /srv/$manifest
                  [ "$?" -ne "0" ]
              do
                  echo "failed to apply /srv/$manifest, retrying in 5 sec"
                  sleep 5s
              done
          done
          echo "Addons successfully installed"
{{ range .Extension.VerbatimSections }}
{{ .Content }}
{{ end }}
`

	WorkerTemplate = `#ignition
ignition:
  version: 2.0.0
systemd:
  units:
    {{range .Extension.Units}}
    - name: {{.Metadata.Name}}
      enable: {{.Metadata.Enable}}
      contents: |
        {{range .Content}}{{.}}
        {{end}}{{end}}
    - name: kubelet.service
      enable: true
      contents: |
        [Service]
        Restart=always
        RestartSec=0
        TimeoutStopSec=10
        EnvironmentFile=/etc/network-environment
        Environment="IMAGE={{.Cluster.Kubernetes.Hyperkube.Docker.Image}}"
        Environment="NAME=%p.service"
        Environment="NETWORK_CONFIG_CONTAINER="
        ExecStartPre=/usr/bin/docker pull $IMAGE
        ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
        ExecStartPre=-/usr/bin/docker rm -f $NAME
        ExecStart=/bin/sh -c "/usr/bin/docker run --rm --pid=host --net=host --privileged=true \
        -v /:/rootfs:ro \
        -v /sys:/sys:ro \
        -v /dev:/dev:rw \
        -v /var/log/pods:/var/log/pods:rw \
        -v /run/calico/:/run/calico/:rw \
        -v /run/docker/:/run/docker/:rw \
        -v /run/docker.sock:/run/docker.sock:rw \
        -v /usr/lib/os-release:/etc/os-release \
        -v /usr/share/ca-certificates/:/etc/ssl/certs \
        -v /var/lib/docker/:/var/lib/docker:rw \
        -v /var/lib/kubelet/:/var/lib/kubelet:rw,rslave \
        -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
        -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
        -v /etc/cni/net.d/:/etc/cni/net.d/ \
        -v /opt/cni/bin/:/opt/cni/bin/ \
        -e ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/etcd/server-ca.pem \
        -e ETCD_CERT_FILE=/etc/kubernetes/ssl/etcd/server-crt.pem \
        -e ETCD_KEY_FILE=/etc/kubernetes/ssl/etcd/server-key.pem \
        --name $NAME \
        $IMAGE \
        /hyperkube kubelet \
        --address=${DEFAULT_IPV4} \
        --port={{.Cluster.Kubernetes.Kubelet.Port}} \
        --hostname-override=${DEFAULT_IPV4} \
        --node-ip=${DEFAULT_IPV4} \
        --api-servers=https://{{.Cluster.Kubernetes.API.Domain}} \
        --containerized \
        --enable-server \
        --logtostderr=true \
        --machine-id-file=/rootfs/etc/machine-id \
        --cadvisor-port=4194 \
        --healthz-bind-address=${DEFAULT_IPV4} \
        --healthz-port=10248 \
        --cluster-dns={{.Cluster.Kubernetes.DNS.IP}} \
        --cluster-domain={{.Cluster.Kubernetes.Domain}} \
        --network-plugin=cni \
        --register-node=true \
        --register-schedulable=false \
        --allow-privileged=true \
        --kubeconfig=/etc/kubernetes/config/kubelet-kubeconfig.yml \
        --node-labels="role=master,kubernetes.io/hostname=${HOSTNAME},ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
        --v=2"
        ExecStop=-/usr/bin/docker stop -t 10 $NAME
        ExecStopPost=-/usr/bin/docker rm -f $NAME
        [Install]
        WantedBy=multi-user.target
    - name: wait-for-domains.service
      enable: true
      contents: |
        [Unit]
        Description=Wait for etcd and k8s API domains to be available

        [Service]
        Type=oneshot
        ExecStart=/opt/wait-for-domains

        [Install]
        WantedBy=multi-user.target
    - name: kubeadm-join.service
      enable: true
      contents: |
        [Unit]
        Description=Initialize master with kubeadm
        Wants=kubelet.service wait-for-domains.service
        [Service]
        Type=oneshot
	Environment=IMAGE=maru/kubeadm
        Environment=NAME=%p.service
        EnvironmentFile=/etc/network-environment
        ExecStartPre=-/usr/bin/docker stop $NAME
        ExecStartPre=-/usr/bin/docker rm $NAME
        ExecStartPre=-/usr/bin/docker pull $IMAGE
        ExecStart=/usr/bin/docker run \
            --name $NAME \
            $IMAGE \
            kubeadm join --token {{.KubeadmToken}}
        [Install]
        WantedBy=multi-user.target
    - name: update-engine.service
      enable: false
    - name: locksmithd.service
      mask: true
    - name: fleet.service
      mask: true
    - name: systemd-networkd-wait-online.service
      enable: true
    - name: docker.service
      enable: true
      dropins:
        - name: 10-giantswarm-extra-args.conf
          contents: |
            [Service]
            Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=cgroupfs {{.Cluster.Docker.Daemon.ExtraArgs}}"
storage:
  files:
    {{ range.Extension.Files}}
    - filesystem: root
      path: {{.Metadata.Path}}
      {{ if .Metadata.Encoding }}
      encoding: {{.Metadata.Encoding}}
      {{ end }}
      mode: {{printf "%#o" .Metadata.Permissions}}
      contents:
        inline: |
          {{range .Content}}{{.}}
          {{end}}{{end}}
    - filesystem: root
      path: /etc/kubernetes/config/kubelet-kubeconfig.yml
      mode: 0664
      contents:
        inline: |
          apiVersion: v1
          kind: Config
          users:
          - name: kubelet
            user:
              client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
              client-key: /etc/kubernetes/ssl/apiserver-key.pem
          clusters:
          - name: local
            cluster:
              certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
          contexts:
          - context:
              cluster: local
              user: kubelet
            name: service-account-context
          current-context: service-account-context
    - filesystem: root
      path: /opt/wait-for-domains
      mode: 0544
      contents:
        inline: |
          #!/bin/bash
          domains="{{.Cluster.Etcd.Domain}} {{.Cluster.Kubernetes.API.Domain}}"

          for domain in $domains; do
              until nslookup $domain; do
                  echo "Waiting for domain $domain to be available"
                  sleep 5
              done

              echo "Successfully resolved domain $domain"
          done
    - filesystem: root
      path: /etc/ssh/sshd_config
      mode: 0600
      contents:
        inline: |
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
{{ range .Extension.VerbatimSections }}
{{ .Content }}
{{ end }}
`

	testTemplate = `foo: {{.Foo}}`
)

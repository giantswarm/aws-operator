package cloudconfig

const (
	MasterTemplate = `#cloud-config
hostname: {{.Node.Hostname}}
write_files:
- path: /etc/hosts
  permissions: 0644
  owner: root
  content: |
    127.0.0.1 localhost
    127.0.0.1 {{.Node.Hostname}}
    127.0.0.1 etcd.giantswarm
    127.0.0.1 {{.Cluster.Kubernetes.API.Domain}}
- path: /etc/resolv.conf
  permissions: 0644
  owner: root
  content: |
    nameserver 8.8.8.8
    nameserver 8.8.4.4
- path: /srv/kubedns-dep.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion:  extensions/v1beta1
    kind: Deployment
    metadata:
      name: kube-dns
      namespace: kube-system
      labels:
        k8s-app: kube-dns
        kubernetes.io/cluster-service: "true"
    spec:
      strategy:
        rollingUpdate:
          maxSurge: 10%
          maxUnavailable: 0
      replicas: 3
      selector:
        matchLabels:
          k8s-app: kube-dns
      template:
        metadata:
          labels:
            k8s-app: kube-dns
            kubernetes.io/cluster-service: "true"
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ''
            scheduler.alpha.kubernetes.io/tolerations: '[{"key":"CriticalAddonsOnly", "operator":"Exists"}]'
        spec:
          containers:
          - name: kubedns
            image: gcr.io/google_containers/kubedns-amd64:1.9
            volumeMounts:
            - name: config
              mountPath: /etc/kubernetes/config/
              readOnly: false
            - name: ssl
              mountPath: /etc/kubernetes/ssl/
              readOnly: false
            resources:
              limits:
                cpu: 100m
                memory: 170Mi
              requests:
                cpu: 100m
                memory: 70Mi
            args:
            # command = "/kube-dns
            - --dns-port=10053
            - --domain={{.Cluster.Kubernetes.Domain}}
            - --kubecfg-file=/etc/kubernetes/config/kubelet-kubeconfig.yml
            - --kube-master-url=https://{{.Cluster.Kubernetes.API.Domain}}
            ports:
            - containerPort: 10053
              name: dns-local
              protocol: UDP
            - containerPort: 10053
              name: dns-tcp-local
              protocol: TCP
            livenessProbe:
              httpGet:
                path: /healthz
                port: 8080
                scheme: HTTP
              initialDelaySeconds: 60
              successThreshold: 1
              failureThreshold: 5
              timeoutSeconds: 5
            readinessProbe:
              httpGet:
                path: /readiness
                port: 8081
                scheme: HTTP
              initialDelaySeconds: 30
              timeoutSeconds: 5
          - name: dnsmasq
            image: gcr.io/google_containers/kube-dnsmasq-amd64:1.4
            args:
            - --cache-size=1000
            - --no-resolv
            - --server=127.0.0.1#10053
            - --log-facility=-
            ports:
            - containerPort: 53
              name: dns
              protocol: UDP
            - containerPort: 53
              name: dns-tcp
              protocol: TCP
            resources:
              requests:
                cpu: 150m
                memory: 10Mi
          - name: healthz
            image: gcr.io/google_containers/exechealthz-amd64:1.2
            resources:
              limits:
                cpu: 10m
                memory: 50Mi
              requests:
                cpu: 10m
                memory: 50Mi
            args:
            - -cmd=nslookup kubernetes.default.svc.{{.Cluster.Kubernetes.Domain}} 127.0.0.1 >/dev/null && nslookup kubernetes.default.svc.{{.Cluster.Kubernetes.Domain}} 127.0.0.1:10053 >/dev/null
            - -port=8080
            - -quiet
            ports:
            - containerPort: 8080
              protocol: TCP
          dnsPolicy: Default  # Don't use cluster DNS.
          volumes:
          - name: config
            hostPath:
              path: /etc/kubernetes/config/
          - name: ssl
            hostPath:
              path: /etc/kubernetes/ssl/
- path: /srv/kubedns-svc.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Service
    metadata:
      name: kube-dns
      namespace: kube-system
      labels:
        k8s-app: kube-dns
        kubernetes.io/cluster-service: "true"
        kubernetes.io/name: "KubeDNS"
    spec:
      selector:
        k8s-app: kube-dns
      clusterIP: {{.Cluster.Kubernetes.DNS.IP}}
      ports:
      - name: dns
        port: 53
        protocol: UDP
      - name: dns-tcp
        port: 53
        protocol: TCP
- path: /srv/calico-system.json
  owner: root
  permissions: 0644
  content: |
    {
      "apiVersion": "v1",
      "kind": "Namespace",
      "metadata": {
        "name": "calico-system"
      }
    }
- path: /srv/network-policy.json
  owner: root
  permissions: 0644
  content: |
    {
      "kind": "ThirdPartyResource",
      "apiVersion": "extensions/v1beta1",
      "metadata": {
        "name": "network-policy.net.alpha.kubernetes.io"
      },
      "description": "Specification for a network isolation policy",
      "versions": [
        {
          "name": "v1alpha1"
        }
      ]
    }
- path: /srv/fallback-server-dep.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: fallback-server
      namespace: kube-system
      labels:
        app: fallback-server
    spec:
      replicas: 2
      template:
        metadata:
          labels:
            app: fallback-server
        spec:
          containers:
          - name: fallback-server
            image: gcr.io/google_containers/defaultbackend:1.2
            args:
            - --port=8000
            readinessProbe:
              httpGet:
                path: /healthz
                port: 8000
                scheme: HTTP
            livenessProbe:
              httpGet:
                path: /healthz
                port: 8000
                scheme: HTTP
              initialDelaySeconds: 10
              timeoutSeconds: 1
- path: /srv/fallback-server-svc.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Service
    metadata:
      name: fallback-server
      namespace: kube-system
      labels:
        app: fallback-server
    spec:
      type: NodePort
      ports:
      - port: 8000
      selector:
        app: fallback-server
- path: /srv/ingress-controller-dep.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: ingress-controller
      namespace: kube-system
      labels:
        app: ingress-controller
    spec:
      replicas: 3
      strategy:
        type: RollingUpdate
        rollingUpdate:
          maxUnavailable: 2
      template:
        metadata:
          labels:
            app: ingress-controller
          annotations:
            scheduler.alpha.kubernetes.io/affinity: >
              {
                "podAntiAffinity": {
                  "preferredDuringSchedulingIgnoredDuringExecution": [
                    {
                      "labelSelector": {
                        "matchExpressions": [
                          {
                            "key": "app",
                            "operator": "In",
                            "values": ["ingress-controller"]
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
          - name: ingress-controller
            image: gcr.io/google_containers/nginx-ingress-controller:0.8.3
            args:
            - /nginx-ingress-controller
            - --default-backend-service=kube-system/fallback-server
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
                port: 80
              initialDelaySeconds: 30
              timeoutSeconds: 1
            livenessProbe:
              httpGet:
                path: /healthz
                port: 80
              initialDelaySeconds: 30
              timeoutSeconds: 1
- path: /srv/ingress-controller-svc.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Service
    metadata:
      name: ingress-controller
      namespace: kube-system
      labels:
        app: ingress-controller
    spec:
      type: NodePort
      ports:
      - name: http
        port: 80
        nodePort: 30010
      - name: https
        port: 443
        nodePort: 30011
      selector:
        app: ingress-controller
- path: /opt/k8s-addons
  permissions: 0544
  content: |
      #!/bin/bash
      # TODO change those to 443
      while ! curl --output /dev/null --silent --head --fail --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem "https://{{.Cluster.Kubernetes.API.Domain}}:6443"; do sleep 1 && echo 'Waiting for master'; done

      echo "K8S: DNS addons"
      curl -H "Content-Type: application/yaml" \
        -XPOST -d"$(cat /srv/kubedns-dep.yaml)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/apis/extensions/v1beta1/namespaces/kube-system/deployments"
      curl -H "Content-Type: application/yaml" \
        -XPOST -d"$(cat /srv/kubedns-svc.yaml)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/api/v1/namespaces/kube-system/services"

      echo "K8S: Calico Policy"
      curl -H "Content-Type: application/json" \
        -XPOST -d"$(cat /srv/calico-system.json)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/api/v1/namespaces/"

      echo "K8S: Fallback Server"
      curl -H "Content-Type: application/yaml" \
        -XPOST -d"$(cat /srv/fallback-server-dep.yml)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/apis/extensions/v1beta1/namespaces/kube-system/deployments"
      curl -H "Content-Type: application/yaml" \
        -XPOST -d"$(cat /srv/fallback-server-svc.yml)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/api/v1/namespaces/kube-system/services"

      echo "K8S: Ingress Controller"
      curl -H "Content-Type: application/yaml" \
        -XPOST -d"$(cat /srv/ingress-controller-dep.yml)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/apis/extensions/v1beta1/namespaces/kube-system/deployments"
      curl -H "Content-Type: application/yaml" \
        -XPOST -d"$(cat /srv/ingress-controller-svc.yml)" \
        --cacert /etc/kubernetes/ssl/apiserver-ca.pem --cert /etc/kubernetes/ssl/apiserver.pem --key /etc/kubernetes/ssl/apiserver-key.pem \
        "https://{{.Cluster.Kubernetes.API.Domain}}:6443/api/v1/namespaces/kube-system/services"

      echo "Addons successfully installed"
- path: /etc/kubernetes/config/controller-manager-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: controller-manager
      user:
        client-certificate: /etc/kubernetes/ssl/apiserver.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
    contexts:
    - context:
        cluster: local
        user: controller-manager
      name: service-account-context
    current-context: service-account-context
- path: /etc/kubernetes/config/scheduler-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: scheduler
      user:
        client-certificate: /etc/kubernetes/ssl/apiserver.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
    contexts:
    - context:
        cluster: local
        user: scheduler
      name: service-account-context
    current-context: service-account-context

- path: /etc/kubernetes/ssl/apiserver-crt.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.APIServerCrt}}

- path: /etc/kubernetes/ssl/apiserver-ca.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.APIServerCACrt}}

- path: /etc/kubernetes/ssl/apiserver-key.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.APIServerKey}}

- path: /etc/kubernetes/ssl/calico/client-crt.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.CalicoClientCrt}}

- path: /etc/kubernetes/ssl/calico/client-ca.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.CalicoClientCACrt}}

- path: /etc/kubernetes/ssl/calico/client-key.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.CalicoClientKey}}

- path: /etc/kubernetes/ssl/etcd/server-crt.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.EtcdServerCrt}}

- path: /etc/kubernetes/ssl/etcd/server-ca.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.EtcdServerCACrt}}

- path: /etc/kubernetes/ssl/etcd/server-key.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.EtcdServerKey}}

{{range .Files}}- path: {{.Metadata.Path}}
  owner: {{.Metadata.Owner}}
  permissions: {{printf "%#o" .Metadata.Permissions}}
  content: |
    {{range .Content}}{{.}}
    {{end}}{{end}}

coreos:
  units:
  - name: set-ownership-etcd-data-dir.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Set ownership to etcd2 data dir
      Wants=network-online.target
      After=etc-kubernetes-data-etcd.mount

      [Service]
      Type=oneshot
      RemainAfterExit=yes
      TimeoutStartSec=0
      ExecStart=/usr/bin/mkdir -p /etc/kubernetes/data/etcd
      ExecStart=/usr/bin/chown etcd:etcd /etc/kubernetes/data/etcd
  - name: docker.service
    enable: true
    command: start
    drop-ins:
    - name: 10-giantswarm-extra-args.conf
      content: |
        [Service]
        Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=systemd {{.Cluster.Docker.Daemon.ExtraArgs}}
  - name: k8s-setup-network-env.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-setup-network-env Service
      Wants=network-online.target docker.service
      After=network-online.target docker.service

      [Service]
      Type=oneshot
      RemainAfterExit=yes
      TimeoutStartSec=0
      Environment="IMAGE={{.Cluster.Operator.NetworkSetup.Docker.Image}}"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/mkdir -p /opt/bin/
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host -v /etc:/etc --name $NAME $IMAGE
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: etcd2.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=etcd2
      Requires=k8s-setup-network-env.service decrypt-tls-assets.service
      After=k8s-setup-network-env.service decrypt-tls-assets.service
      Conflicts=etcd.service
      Wants=calico-node.service
      StartLimitIntervalSec=0

      [Service]
      User=etcd
      Type=notify
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      LimitNOFILE=40000
      EnvironmentFile=/etc/network-environment
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/etcd/server-ca.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/etcd/server-ca.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/etcd/server-crt.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/etcd/server-crt.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/etcd/server-key.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/etcd/server-key.pem to be written' && sleep 1; done"

      # TODO, switch to {{.Cluster.Etcd.Domain}}:443 when the ingress controllers are set up
      ExecStart=/usr/bin/etcd2 --advertise-client-urls=https://{{.Cluster.Etcd.Domain}}:2379,http://127.0.0.1:2383 \
                               --data-dir=/etc/kubernetes/data/etcd/ \
                               --initial-advertise-peer-urls=https://{{.Cluster.Etcd.Domain}}:2379 \
                               --listen-client-urls=https://0.0.0.0:2379,http://127.0.0.1:2383 \
                               --listen-peer-urls=https://{{.Cluster.Etcd.Domain}}:2380 \
                               --initial-cluster-token k8s-etcd-cluster \
                               --initial-cluster etcd0=https://{{.Cluster.Etcd.Domain}}:2379 \
                               --initial-cluster-state new \
                               --ca-file=/etc/kubernetes/ssl/etcd/server-ca.pem \
                               --cert-file=/etc/kubernetes/ssl/etcd/server-crt.pem \
                               --key-file=/etc/kubernetes/ssl/etcd/server-key.pem \
                               --peer-ca-file=/etc/kubernetes/ssl/etcd/server-ca.pem \
                               --peer-cert-file=/etc/kubernetes/ssl/etcd/server-crt.pem \
                               --peer-key-file=/etc/kubernetes/ssl/etcd/server-key.pem \
                               --peer-client-cert-auth=true \
                               --name etcd0

      [Install]
      WantedBy=multi-user.target
  - name: etcd2-restart.service
    enable: true
    content: |
      [Unit]
      Description=etcd2-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop etcd2.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet etcd2.service; do sleep 1 && echo waiting for etcd2 to stop; done'
      ExecStart=/usr/bin/systemctl start etcd2.service

      [Install]
      WantedBy=multi-user.target
  - name: etcd2-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=13:00
      Unit=etcd2-restart.service

      [Install]
      WantedBy=multi-user.target
  # TODO(nhlfr): Set up Calico on Kubernetes, in example by http://docs.projectcalico.org/v2.0/getting-started/kubernetes/installation/hosted/kubeadm/calico.yaml.
  # Or at least use anything which doesn't download binaries in systemd unit...
  - name: calico-node.service
    runtime: true
    command: start
    content: |
      [Unit]
      Description=Calico per-host agent
      Requires=etcd2.service decrypt-tls-assets.service
      After=etcd2.service decrypt-tls-assets.service
      Wants=k8s-api-server.service k8s-controller-manager.service k8s-scheduler.service
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="ETCD_AUTHORITY=127.0.0.1:2383"
      ExecStartPre=/usr/bin/wget -O /opt/bin/calicoctl https://s3-eu-west-1.amazonaws.com/downloads.giantswarm.io/calicoctl/v0.22.0/calicoctl
      ExecStartPre=/usr/bin/chmod +x /opt/bin/calicoctl
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/calico/client-ca.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/calico/client-ca.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/calico/client-crt.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/calico/client.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/calico/client-key.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/calico/client-key.pem to be written' && sleep 1; done"
      # TODO remove 2379 when we use ingress controllers
      ExecStartPre=/bin/bash -c "while ! curl --output /dev/null --silent --fail --cacert /etc/kubernetes/ssl/calico/client-ca.pem --cert /etc/kubernetes/ssl/calico/client-crt.pem --key /etc/kubernetes/ssl/calico/client-key.pem https://{{.Cluster.Etcd.Domain}}:2379/version; do sleep 1 && echo 'Waiting for etcd master to be responsive'; done"
      ExecStartPre=/opt/bin/calicoctl pool add {{.Cluster.Calico.Subnet}}/{{.Cluster.Calico.CIDR}} --ipip --nat-outgoing
      ExecStart=/opt/bin/calicoctl node --ip=${DEFAULT_IPV4}  --detach=false --node-image=giantswarm/node:v0.22.0
      ExecStop=/opt/bin/calicoctl node stop --force
      ExecStopPost=/bin/bash -c "find /tmp/ -name '_MEI*' | xargs -I {} rm -rf {}"

      [Install]
      WantedBy=multi-user.target
  - name: calico-node-restart.service
    enable: true
    content: |
      [Unit]
      Description=calico-node-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop calico-node.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet calico-node.service; do sleep 1 && echo waiting for calico-node to stop; done'
      ExecStart=/usr/bin/systemctl start calico-node.service

      [Install]
      WantedBy=multi-user.target
  - name: calico-node-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=*-01,04,07,10-01 14:00:00
      Unit=calico-node-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: update-engine.service
    enable: false
    command: stop
    mask: true
  - name: locksmithd.service
    command: stop
    mask: true
  - name: fleet.service
    mask: true
    command: stop
  - name: systemd-networkd-wait-online.service
    enable: true
    command: start
  - name: k8s-api-server.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-api-server
      Requires=calico-node.service k8s-key-generator.service
      After=calico-node.service k8s-key-generator.service
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE={{.Cluster.Kubernetes.Hyperkube.Docker.Image}}"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/mkdir -p /etc/kubernetes/manifests
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      # TODO figure out this
      ExecStartPre=/usr/bin/ln -sf /etc/kubernetes/ssl/apiserver-crt.pem /etc/kubernetes/ssl/apiserver.pem
      # TODO change 0.0.0.0 to ${DEFAULT_IP}
      # TODO change 2379 to 443
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/secrets/token_sign_key.pem ]; do echo 'Waiting for /etc/kubernetes/secrets/token_sign_key.pem to be written' && sleep 1; done"
      ExecStart=/usr/bin/docker run --rm --name $NAME --net=host \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/secrets/token_sign_key.pem:/etc/kubernetes/secrets/token_sign_key.pem \
      $IMAGE \
      /hyperkube apiserver \
      --allow_privileged=true \
      --runtime_config=api/v1 \
      --insecure_bind_address=0.0.0.0 \
      --insecure_port={{.Cluster.Kubernetes.API.InsecurePort}} \
      --kubelet_https=true \
      --secure_port={{.Cluster.Kubernetes.API.SecurePort}} \
      --bind-address=0.0.0.0 \
      --etcd-prefix={{.Cluster.Etcd.Prefix}} \
      --admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota \
      --service-cluster-ip-range={{.Cluster.Kubernetes.API.ClusterIPRange}} \
      --etcd_servers=https://{{.Cluster.Etcd.Domain}}:2379 \
      --etcd-cafile=/etc/kubernetes/ssl/etcd/server-ca.pem \
      --etcd-certfile=/etc/kubernetes/ssl/etcd/server-crt.pem \
      --etcd-keyfile=/etc/kubernetes/ssl/etcd/server-key.pem \
      --advertise-address=${DEFAULT_IPV4} \
      --runtime-config=extensions/v1beta1/deployments=true,extensions/v1beta1/daemonsets=true,extensions/v1beta1=true,extensions/v1beta1/thirdpartyresources=true,extensions/v1beta1/networkpolicies=true,batch/v2alpha1 \
      --logtostderr=true \
      --tls-cert-file=/etc/kubernetes/ssl/apiserver-crt.pem \
      --tls-private-key-file=/etc/kubernetes/ssl/apiserver-key.pem \
      --client-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem \
      --service-account-key-file=/etc/kubernetes/secrets/token_sign_key.pem
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME

  - name: k8s-key-generator.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-key-generator Service

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/mkdir -p /etc/kubernetes/secrets
      ExecStartPre=/usr/bin/rm -rf /etc/kubernetes/secrets/token_sign_key.pem
      ExecStart=/usr/bin/openssl genrsa -out /etc/kubernetes/secrets/token_sign_key.pem 2048

      [Install]
      WantedBy=multi-user.target
  - name: k8s-api-server-restart.service
    enable: true
    content: |
      [Unit]
      Description=k8s-api-server-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop k8s-api-server.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet k8s-api-server.service; do sleep 1 && echo waiting for k8s-api-server to stop; done'
      ExecStart=/usr/bin/systemctl start k8s-api-server.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-api-server-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=15:00
      Unit=k8s-api-server-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-controller-manager.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-controller-manager Service
      Requires=calico-node.service k8s-key-generator.service
      After=calico-node.service k8s-key-generator.service
      StartLimitIntervalSec=0

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
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/secrets/token_sign_key.pem ]; do echo 'Waiting for /etc/kubernetes/secrets/token_sign_key.pem to be written' && sleep 1; done"
      ExecStart=/usr/bin/docker run --rm --net=host --name $NAME \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      -v /etc/kubernetes/secrets/token_sign_key.pem:/etc/kubernetes/secrets/token_sign_key.pem \
      $IMAGE \
      /hyperkube controller-manager \
      --master=https://{{.Cluster.Kubernetes.API.Domain}}:6443 \
      --logtostderr=true \
      --v=2 \
      --kubeconfig=/etc/kubernetes/config/controller-manager-kubeconfig.yml \
      --root-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem \
      --service-account-private-key-file=/etc/kubernetes/secrets/token_sign_key.pem
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-controller-manager-restart.service
    enable: true
    content: |
      [Unit]
      Description=k8s-controller-manager-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop k8s-controller-manager.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet k8s-controller-manager.service; do sleep 1 && echo waiting for k8s-controller-manager to stop; done'
      ExecStart=/usr/bin/systemctl start k8s-controller-manager.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-controller-manager-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=15:00
      Unit=k8s-controller-manager-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-scheduler.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-scheduler Service
      Requires=calico-node.service
      After=calico-node.service
      StartLimitIntervalSec=0

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
      # TODO change 6443 to 443
      ExecStart=/usr/bin/docker run --rm --net=host --name $NAME \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      $IMAGE \
      /hyperkube scheduler \
      --master=https://{{.Cluster.Kubernetes.API.Domain}}:6443 \
      --logtostderr=true \
      --v=2 \
      --kubeconfig=/etc/kubernetes/config/scheduler-kubeconfig.yml
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-scheduler-restart.service
    enable: true
    content: |
      [Unit]
      Description=k8s-scheduler-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop k8s-scheduler.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet k8s-scheduler.service; do sleep 1 && echo waiting for k8s-scheduler to stop; done'
      ExecStart=/usr/bin/systemctl start k8s-scheduler.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-scheduler-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=15:00
      Unit=k8s-scheduler-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-addons.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Kubernetes Addons
      Wants=k8s-scheduler.service
      After=k8s-scheduler.service
      [Service]
      Type=oneshot
      EnvironmentFile=/etc/network-environment
      ExecStart=/opt/k8s-addons
      [Install]
      WantedBy=multi-user.target
  - name: k8s-policy-controller.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-policy-controller Service
      Wants=k8s-api-server.service
      Requires=k8s-addons.service
      After=k8s-addons.service

      [Service]
      Restart=always
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE={{.Cluster.Docker.ImageNamespace}}/k8s-policy-controller:v0.2.0"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host \
      --name $NAME \
      -e ETCD_ENDPOINTS=http://127.0.0.1:2383 \
      -e K8S_API=http://localhost:{{.Cluster.Kubernetes.API.InsecurePort}} \
      -e LEADER_ELECTION=true \
      $IMAGE
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: leader-elector.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=leader-elector Service
      Requires=k8s-policy-controller.service
      After=k8s-policy-controller.service

      [Service]
      Restart=always
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE={{.Cluster.Docker.ImageNamespace}}/leader-elector:v0.1.0"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host \
      --name $NAME \
      $IMAGE \
      --election=calico-policy-election \
      --election-namespace=calico-system \
      --http=127.0.0.1:4040
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: node-exporter.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Prometheus Node Exporter Service
      Requires=docker.service
      After=docker.service

      [Service]
      Restart=always
      Environment="IMAGE=prom/node-exporter:0.12.0"
      Environment="NAME=%p.service"
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm \
        -p 91:91 \
        --net=host \
        --name $NAME \
        $IMAGE \
        --web.listen-address=:91
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME

      [Install]
      WantedBy=multi-user.target
  {{range .Units}}- name: {{.Metadata.Name}}
    enable: {{.Metadata.Enable}}
    command: {{.Metadata.Command}}
    content: |
      {{range .Content}}{{.}}
      {{end}}{{end}}
  update:
    reboot-strategy: off
`

	WorkerTemplate = `#cloud-config
hostname: {{.Node.Hostname}}
write_files:
- path: /etc/hosts
  permissions: 0644
  owner: root
  content: |
    127.0.0.1 localhost
    127.0.0.1 {{.Node.Hostname}}
    127.0.0.1 etcd.giantswarm
    {{.PrivateIP}} {{.Cluster.Kubernetes.API.Domain}}
- path: /srv/10-calico.conf
  owner: root
  permissions: 0755
  content: |
    {
        "name": "calico-k8s-network",
        "type": "calico",
        "etcd_endpoints": "https://{{.Cluster.Etcd.Domain}}:443",
        "log_level": "info",
        "ipam": {
            "type": "calico-ipam"
        },
        "mtu": {{.Cluster.Calico.MTU}},
        "policy": {
            "type": "k8s",
            "k8s_api_root": "https://{{.Cluster.Kubernetes.API.Domain}}/api/v1/",
            "k8s_client_certificate": "/etc/kubernetes/ssl/calico/client.pem",
            "k8s_client_key": "/etc/kubernetes/ssl/calico/client-key.pem",
            "k8s_certificate_authority": "/etc/kubernetes/ssl/calico/client-ca.pem"
        }
    }
- path: /etc/resolv.conf
  permissions: 0644
  owner: root
  content: |
    nameserver 8.8.8.8
    nameserver 8.8.4.4
- path: /etc/kubernetes/config/proxy-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: proxy
      user:
        client-certificate: /etc/kubernetes/ssl/worker.pem
        client-key: /etc/kubernetes/ssl/worker-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/worker-ca.pem
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
        client-certificate: /etc/kubernetes/ssl/worker.pem
        client-key: /etc/kubernetes/ssl/worker-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/worker-ca.pem
    contexts:
    - context:
        cluster: local
        user: kubelet
      name: service-account-context
    current-context: service-account-context

- path: /etc/kubernetes/ssl/apiserver-crt.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.APIServerCrt}}

- path: /etc/kubernetes/ssl/apiserver-ca.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.APIServerCACrt}}

- path: /etc/kubernetes/ssl/apiserver-key.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.APIServerKey}}

- path: /etc/kubernetes/ssl/calico/client-crt.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.CalicoClientCrt}}

- path: /etc/kubernetes/ssl/calico/client-ca.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.CalicoClientCACrt}}

- path: /etc/kubernetes/ssl/calico/client-key.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.CalicoClientKey}}

- path: /etc/kubernetes/ssl/etcd/server-crt.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.EtcdServerCrt}}

- path: /etc/kubernetes/ssl/etcd/server-ca.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.EtcdServerCACrt}}

- path: /etc/kubernetes/ssl/etcd/server-key.pem.enc
  encoding: gzip+base64
  content: {{.TLSAssets.EtcdServerKey}}

{{range .Files}}- path: {{.Metadata.Path}}
  owner: {{.Metadata.Owner}}
  permissions: {{printf "%#o" .Metadata.Permissions}}
  content: |
    {{range .Content}}{{.}}
    {{end}}{{end}}

coreos:
  units:
  - name: update-engine.service
    enable: false
    command: stop
    mask: true
  - name: locksmithd.service
    command: stop
    mask: true
  - name: etcd2.service
    enable: true
    command: start
  - name: fleet.service
    command: stop
    mask: true
  - name: systemd-networkd-wait-online.service
    enable: true
    command: start
  - name: docker.service
    enable: true
    command: start
    drop-ins:
    - name: 10-giantswarm-extra-args.conf
      content: |
        [Service]
        Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=systemd {{.Cluster.Docker.Daemon.ExtraArgs}}
  - name: k8s-setup-network-env.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-setup-network-env Service
      Wants=network-online.target docker.service
      After=network-online.target docker.service

      [Service]
      Type=oneshot
      RemainAfterExit=yes
      TimeoutStartSec=0
      Environment="IMAGE={{.Cluster.Operator.NetworkSetup.Docker.Image}}"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/mkdir -p /etc/kubernetes/cni/net.d/
      ExecStartPre=-/usr/bin/cp /srv/10-calico.conf /etc/kubernetes/cni/net.d/10-calico.conf
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host -v /etc:/etc --name $NAME $IMAGE
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  # TODO(nhlfr): Set up Calico on Kubernetes, in example by http://docs.projectcalico.org/v2.0/getting-started/kubernetes/installation/hosted/kubeadm/calico.yaml.
  # Or at least use anything which doesn't download binaries in systemd unit...
  - name: calico-node.service
    runtime: true
    command: start
    content: |
      [Unit]
      Description=calicoctl node
      Requires=k8s-setup-network-env.service decrypt-tls-assets.service
      After=k8s-setup-network-env.service decrypt-tls-assets.service
      Wants=k8s-proxy.service k8s-kubelet.service
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      StartLimitIntervalSec=0
      EnvironmentFile=/etc/environment
      Environment="ETCD_AUTHORITY={{.Cluster.Kubernetes.API.Domain}}:2379"
      Environment="ETCD_SCHEME=https"
      Environment="ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/calico/client-ca.pem"
      Environment="ETCD_CERT_FILE=/etc/kubernetes/ssl/calico/client-crt.pem"
      Environment="ETCD_KEY_FILE=/etc/kubernetes/ssl/calico/client-key.pem"
      ExecStartPre=/usr/bin/mkdir -p /opt/cni/bin
      ExecStartPre=/usr/bin/wget -O /opt/cni/bin/calico https://s3-eu-west-1.amazonaws.com/downloads.giantswarm.io/calico-cni/v1.4.2/calico
      ExecStartPre=/usr/bin/chmod +x /opt/cni/bin/calico
      ExecStartPre=/usr/bin/wget -O /opt/cni/bin/calico-ipam https://s3-eu-west-1.amazonaws.com/downloads.giantswarm.io/calico-cni/v1.4.2/calico-ipam
      ExecStartPre=/usr/bin/chmod +x /opt/cni/bin/calico-ipam
      ExecStartPre=/usr/bin/mkdir -p /opt/bin/
      ExecStartPre=/usr/bin/wget -O /opt/bin/calicoctl https://s3-eu-west-1.amazonaws.com/downloads.giantswarm.io/calicoctl/v0.22.0/calicoctl
      ExecStartPre=/usr/bin/chmod +x /opt/bin/calicoctl
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/calico/client-ca.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/calico/client-ca.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/calico/client-crt.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/calico/client-crt.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/calico/client-key.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/calico/client-key.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while ! curl --output /dev/null --silent --fail --cacert /etc/kubernetes/ssl/calico/client-ca.pem --cert /etc/kubernetes/ssl/calico/client-crt.pem --key /etc/kubernetes/ssl/calico/client-key.pem https://{{.Cluster.Etcd.Domain}}:2379/version; do sleep 1 && echo 'Waiting for etcd master to be responsive'; done"
      ExecStart=/opt/bin/calicoctl node --ip=${COREOS_PRIVATE_IPV4} --detach=false --node-image=giantswarm/node:v0.22.0
      ExecStop=/opt/bin/calicoctl node stop --force
      ExecStopPost=/bin/bash -c "find /tmp/ -name '_MEI*' | xargs -I {} rm -rf {}"

      [Install]
      WantedBy=multi-user.target
  - name: calico-node-restart.service
    enable: true
    content: |
      [Unit]
      Description=calico-node-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop calico-node.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet calico-node.service; do sleep 1 && echo waiting for calico-node to stop; done'
      ExecStart=/usr/bin/systemctl start calico-node.service

      [Install]
      WantedBy=multi-user.target
  - name: calico-node-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=*-01,04,07,10-01 14:00:00
      Unit=calico-node-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-proxy.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-proxy
      Requires=calico-node.service
      After=calico-node.service
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE={{.Cluster.Kubernetes.Hyperkube.Docker.Image}}
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/worker-ca.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/worker-ca.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/worker.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/worker.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/bash -c "while [ ! -f /etc/kubernetes/ssl/worker-key.pem ]; do echo 'Waiting for /etc/kubernetes/ssl/worker-key.pem to be written' && sleep 1; done"
      ExecStartPre=/bin/sh -c "while ! curl --output /dev/null --silent --head --fail --cacert /etc/kubernetes/ssl/worker-ca.pem --cert /etc/kubernetes/ssl/worker.pem --key /etc/kubernetes/ssl/worker-key.pem https://{{.Cluster.Kubernetes.API.Domain}}; do sleep 1 && echo 'Waiting for master'; done"
      ExecStart=/bin/sh -c "/usr/bin/docker run --rm --net=host --privileged=true \
      --name $NAME \
      -v /usr/share/ca-certificates:/etc/ssl/certs \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      $IMAGE \
      /hyperkube proxy \
      --master=https://{{.Cluster.Kubernetes.API.Domain}} \
      --proxy-mode=iptables \
      --logtostderr=true \
      --kubeconfig=/etc/kubernetes/config/proxy-kubeconfig.yml \
      --v=2"
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-proxy-restart.service
    enable: true
    content: |
      [Unit]
      Description=k8s-proxy-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop k8s-proxy.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet k8s-proxy.service; do sleep 1 && echo waiting for k8s-proxy to stop; done'
      ExecStart=/usr/bin/systemctl start k8s-proxy.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-proxy-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=15:00
      Unit=k8s-proxy-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: k8s-kubelet.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-kubelet
      Requires=calico-node.service
      After=calico-node.service
      StartLimitIntervalSec=0

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
      -v /run/calico/:/run/calico/:rw \
      -v /run/docker/:/run/docker/:rw \
      -v /run/docker.sock:/run/docker.sock:rw \
      -v /usr/lib/os-release:/etc/os-release \
      -v /usr/share/ca-certificates/:/etc/ssl/certs \
      -v /var/lib/docker/:/var/lib/docker:rw \
      -v /var/lib/kubelet/:/var/lib/kubelet:rw,rslave \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      -v /etc/kubernetes/cni/:/etc/kubernetes/cni/ \
      -v /opt/cni/bin/calico:/opt/cni/bin/calico \
      -v /opt/cni/bin/calico-ipam:/opt/cni/bin/calico-ipam \
      -e ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/etcd/client-ca.pem \
      -e ETCD_CERT_FILE=/etc/kubernetes/ssl/etcd/client.pem \
      -e ETCD_KEY_FILE=/etc/kubernetes/ssl/etcd/client-key.pem \
      --name $NAME \
      $IMAGE \
      /hyperkube kubelet \
      --address=${DEFAULT_IPV4} \
      --port=10250 \
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
      --cluster-dns={{.Cluster.Kubernetes.DNS}} \
      --cluster-domain={{.Cluster.Kubernetes.API.Domain}} \
      --network-plugin-dir=/etc/kubernetes/cni/net.d \
      --network-plugin=cni \
      --register-node=true \
      --allow-privileged=true \
      --kubeconfig=/etc/kubernetes/config/kubelet-kubeconfig.yml \
      --node-labels="kubernetes.io/hostname={{.Node.Hostname}},ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
      --v=2"
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-kubelet-restart.service
    enable: true
    content: |
      [Unit]
      Description=k8s-kubelet-restart

      [Service]
      Type=oneshot
      ExecStartPre=/usr/bin/systemctl stop k8s-kubelet.service
      ExecStartPre=/usr/bin/bash -c 'while systemctl is-active --quiet k8s-kubelet.service; do sleep 1 && echo waiting for k8s-kubelet to stop; done'
      ExecStart=/usr/bin/systemctl start k8s-kubelet.service

      [Install]
      WantedBy=multi-user.target
  - name: kubelet-restart.timer
    enable: true
    command: start
    content: |
      [Unit]
      Description=Timer

      [Timer]
      OnCalendar=15:00
      Unit=k8s-kubelet-restart.service

      [Install]
      WantedBy=multi-user.target
  - name: node-exporter.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Prometheus Node Exporter Service
      Requires=docker.service
      After=docker.service

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      Environment="IMAGE=prom/node-exporter:0.12.0"
      Environment="NAME=%p.service"
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm \
        -p 91:91 \
        --net=host \
        --name $NAME \
        $IMAGE \
        --web.listen-address=:91
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME

      [Install]
      WantedBy=multi-user.target
  - name: decrypt-tls-certs.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Decrypt TLS certificates

      [Service]
      ExecStart=/opt/bin/decrypt-tls-assets
  {{range .Units}}- name: {{.Metadata.Name}}
    enable: {{.Metadata.Enable}}
    command: {{.Metadata.Command}}
    content: |
      {{range .Content}}{{.}}
      {{end}}{{end}}
  update:
    reboot-strategy: off
`

	testTemplate = `foo: {{.Foo}}`
)

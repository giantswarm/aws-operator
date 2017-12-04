package v_2_0_0

const MasterTemplate = `#cloud-config
users:
{{ range $index, $user := .Cluster.Kubernetes.SSH.UserList }}  - name: {{ $user.Name }}
    groups:
      - "sudo"
      - "docker"
    ssh-authorized-keys:
       - "{{ $user.PublicKey }}"
{{end}}
write_files:
- path: /srv/calico-ipip-pinger-sa.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: calico-ipip-pinger
      namespace: kube-system
- path: /srv/calico-ipip-pinger.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: extensions/v1beta1
    kind: DaemonSet
    metadata:
      labels:
        app: calico-ipip-pinger
      name: calico-ipip-pinger
      namespace: kube-system
    spec:
      updateStrategy:
        type: RollingUpdate
        rollingUpdate:
          maxUnavailable: 1
      template:
        metadata:
          labels:
            app: calico-ipip-pinger
        spec:
          serviceAccountName: calico-ipip-pinger
          containers:
          - name: calico-ipip-pinger
            image: quay.io/giantswarm/calico-ipip-pinger:0f197abe866c7c811e42534739148345cb3bded5
            imagePullPolicy: Always
            securityContext:
              privileged: true
            env:
              # The location of the Calico etcd cluster.
              - name: ETCD_ENDPOINTS
                valueFrom:
                  configMapKeyRef:
                    name: calico-config
                    key: etcd_endpoints
              # Location of the CA certificate for etcd.
              - name: ETCD_CA_CERT_FILE
                valueFrom:
                  configMapKeyRef:
                    name: calico-config
                    key: etcd_ca
              # Location of the client key for etcd.
              - name: ETCD_KEY_FILE
                valueFrom:
                  configMapKeyRef:
                    name: calico-config
                    key: etcd_key
              # Location of the client certificate for etcd.
              - name: ETCD_CERT_FILE
                valueFrom:
                  configMapKeyRef:
                    name: calico-config
                    key: etcd_cert
            volumeMounts:
              # Mount in the etcd TLS secrets.
              - mountPath: /etc/kubernetes/ssl/etcd
                name: etcd-certs
          volumes:
            # Mount in the etcd TLS secrets.
            - name: etcd-certs
              hostPath:
                path: /etc/kubernetes/ssl/etcd
          tolerations:
          - effect: NoSchedule
            key: node-role.kubernetes.io/master
            operator: Exists
          hostNetwork: true
          restartPolicy: Always
- path: /srv/calico-node-controller-sa.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: calico-node-controller
      namespace: kube-system
- path: /srv/calico-node-controller.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: calico-node-controller
      namespace: kube-system
      labels:
        k8s-app: calico-node-controller
    spec:
      replicas: 1
      strategy:
        type: Recreate
      template:
        metadata:
          name: calico-node-controller
          namespace: kube-system
          labels:
            k8s-app: calico-node-controller
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ''
        spec:
          tolerations:
          - key: node-role.kubernetes.io/master
            operator: Exists
            effect: NoSchedule
          - key: CriticalAddonsOnly
            operator: Exists
          nodeSelector:
            node-role.kubernetes.io/master: ""
          hostNetwork: true
          serviceAccountName: calico-node-controller
          containers:
            - name: calico-node-controller
              image: quay.io/giantswarm/calico-node-controller
              env:
                # The location of the Calico etcd cluster.
                - name: ETCD_ENDPOINTS
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_endpoints
                # Location of the CA certificate for etcd.
                - name: ETCD_CA_CERT_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_ca
                # Location of the client key for etcd.
                - name: ETCD_KEY_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_key
                # Location of the client certificate for etcd.
                - name: ETCD_CERT_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_cert
                # The location of the Kubernetes API.  Use the default Kubernetes
                # service for API access.
                - name: K8S_API
                  value: "https://kubernetes.default:443"
              volumeMounts:
                # Mount in the etcd TLS secrets.
                - mountPath: /etc/kubernetes/ssl/etcd
                  name: etcd-certs
          volumes:
            # Mount in the etcd TLS secrets.
            - name: etcd-certs
              hostPath:
                path: /etc/kubernetes/ssl/etcd
- path: /srv/calico-kube-controllers-sa.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
- path: /srv/calico-node-sa.yaml
  owner: root
  permissions: 644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: calico-node
      namespace: kube-system
- path: /srv/calico-configmap.yaml
  owner: root
  permissions: 644
  content: |
    # Calico Version v2.6.2
    # https://docs.projectcalico.org/v2.6/releases#v2.6.2
    # This manifest includes the following component versions:
    #   calico/node:v2.6.2
    #   calico/cni:v1.11.0
    #   calico/kube-controllers:v1.0.0

    # This ConfigMap is used to configure a self-hosted Calico installation.
    kind: ConfigMap
    apiVersion: v1
    metadata:
      name: calico-config
      namespace: kube-system
    data:
      # Configure this with the location of your etcd cluster.
      etcd_endpoints: "https://{{ .Cluster.Etcd.Domain }}:443"

      # Configure the Calico backend to use.
      calico_backend: "bird"

      # The CNI network configuration to install on each node.
      # TODO: Do we still need to set MTU manually?
      cni_network_config: |-
        {
            "name": "k8s-pod-network",
            "cniVersion": "0.1.0",
            "type": "calico",
            "etcd_endpoints": "__ETCD_ENDPOINTS__",
            "etcd_key_file": "__ETCD_KEY_FILE__",
            "etcd_cert_file": "__ETCD_CERT_FILE__",
            "etcd_ca_cert_file": "__ETCD_CA_CERT_FILE__",
            "log_level": "info",
            "ipam": {
                "type": "calico-ipam"
            },
            "mtu": {{.Cluster.Calico.MTU}},
            "policy": {
                "type": "k8s",
                "k8s_api_root": "https://__KUBERNETES_SERVICE_HOST__:__KUBERNETES_SERVICE_PORT__",
                "k8s_auth_token": "__SERVICEACCOUNT_TOKEN__"
            },
            "kubernetes": {
                "kubeconfig": "__KUBECONFIG_FILEPATH__"
            }
        }

      etcd_ca: "/etc/kubernetes/ssl/etcd/client-ca.pem"
      etcd_cert: "/etc/kubernetes/ssl/etcd/client-crt.pem"
      etcd_key: "/etc/kubernetes/ssl/etcd/client-key.pem"

- path: /srv/calico-ds.yaml
  owner: root
  permissions: 644
  content: |

    # This manifest installs the calico/node container, as well
    # as the Calico CNI plugins and network config on
    # each master and worker node in a Kubernetes cluster.
    kind: DaemonSet
    apiVersion: extensions/v1beta1
    metadata:
      name: calico-node
      namespace: kube-system
      labels:
        k8s-app: calico-node
    spec:
      selector:
        matchLabels:
          k8s-app: calico-node
      updateStrategy:
        type: RollingUpdate
        rollingUpdate:
          maxUnavailable: 1
      template:
        metadata:
          labels:
            k8s-app: calico-node
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ''
        spec:
          # Tolerations part was taken from calico manifest for kubeadm as we are using same taint for master.
          tolerations:
          - key: node-role.kubernetes.io/master
            operator: Exists
            effect: NoSchedule
          - key: CriticalAddonsOnly
            operator: Exists
          hostNetwork: true
          serviceAccountName: calico-node
          containers:
            # Runs calico/node container on each Kubernetes node.  This
            # container programs network policy and routes on each
            # host.
            - name: calico-node
              image: quay.io/calico/node:v2.6.2
              env:
                # The location of the Calico etcd cluster.
                - name: ETCD_ENDPOINTS
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_endpoints
                # Choose the backend to use.
                - name: CALICO_NETWORKING_BACKEND
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: calico_backend
                # Cluster type to identify the deployment type
                - name: CLUSTER_TYPE
                  value: "k8s,bgp"
                # Disable file logging so kubectl logs works.
                - name: CALICO_DISABLE_FILE_LOGGING
                  value: "true"
                # Set Felix endpoint to host default action to ACCEPT.
                - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
                  value: "ACCEPT"
                # Configure the IP Pool from which Pod IPs will be chosen.
                - name: CALICO_IPV4POOL_CIDR
                  value: "{{.Cluster.Calico.Subnet}}/{{.Cluster.Calico.CIDR}}"
                - name: CALICO_IPV4POOL_IPIP
                  value: "always"
                # Disable IPv6 on Kubernetes.
                - name: FELIX_IPV6SUPPORT
                  value: "false"
                # Set Felix logging to "info"
                - name: FELIX_LOGSEVERITYSCREEN
                  value: "info"
                # Set MTU for tunnel device used if ipip is enabled
                - name: FELIX_IPINIPMTU
                  value: "1440"
                # Location of the CA certificate for etcd.
                - name: ETCD_CA_CERT_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_ca
                # Location of the client key for etcd.
                - name: ETCD_KEY_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_key
                # Location of the client certificate for etcd.
                - name: ETCD_CERT_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_cert
                # Auto-detect the BGP IP address.
                - name: IP
                  value: ""
                - name: FELIX_HEALTHENABLED
                  value: "true"
              securityContext:
                privileged: true
              resources:
                requests:
                  cpu: 250m
              livenessProbe:
                httpGet:
                  path: /liveness
                  port: 9099
                periodSeconds: 10
                initialDelaySeconds: 10
                failureThreshold: 6
              readinessProbe:
                httpGet:
                  path: /readiness
                  port: 9099
                periodSeconds: 10
              volumeMounts:
                - mountPath: /lib/modules
                  name: lib-modules
                  readOnly: true
                - mountPath: /var/run/calico
                  name: var-run-calico
                  readOnly: false
                - mountPath: /etc/kubernetes/ssl/etcd
                  name: etcd-certs
            # This container installs the Calico CNI binaries
            # and CNI network config file on each node.
            - name: install-cni
              image: quay.io/calico/cni:v1.11.0
              command: ["/install-cni.sh"]
              env:
                # The location of the Calico etcd cluster.
                - name: ETCD_ENDPOINTS
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_endpoints
                # The CNI network config to install on each node.
                - name: CNI_NETWORK_CONFIG
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: cni_network_config
                # Location of the CA certificate for etcd.
                - name: CNI_CONF_ETCD_CA
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_ca
                # Location of the client key for etcd.
                - name: CNI_CONF_ETCD_KEY
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_key
                # Location of the client certificate for etcd.
                - name: CNI_CONF_ETCD_CERT
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_cert
              volumeMounts:
                - mountPath: /host/opt/cni/bin
                  name: cni-bin-dir
                - mountPath: /host/etc/cni/net.d
                  name: cni-net-dir
                - mountPath: /etc/kubernetes/ssl/etcd
                  name: etcd-certs
          volumes:
            # Used by calico/node.
            - name: etcd-certs
              hostPath:
                path: /etc/kubernetes/ssl/etcd
            - name: lib-modules
              hostPath:
                path: /lib/modules
            - name: var-run-calico
              hostPath:
                path: /var/run/calico
            - name: cni-bin-dir
              hostPath:
                path: /opt/cni/bin
            - name: cni-net-dir
              hostPath:
               path: /etc/cni/net.d
- path: /srv/calico-kube-controllers.yaml
  owner: root
  permissions: 644
  content: |
    # This manifest deploys the Calico Kubernetes controllers.
    # See https://github.com/projectcalico/kube-controllers
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
      labels:
        k8s-app: calico-kube-controllers
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      # The policy controller can only have a single active instance.
      replicas: 1
      strategy:
        type: Recreate
      template:
        metadata:
          name: calico-kube-controllers
          namespace: kube-system
          labels:
            k8s-app: calico-kube-controllers
        spec:
          tolerations:
          - key: node-role.kubernetes.io/master
            operator: Exists
            effect: NoSchedule
          - key: CriticalAddonsOnly
            operator: Exists
          # The controllers must run in the host network namespace so that
          # it isn't governed by policy that would prevent it from working.
          hostNetwork: true
          serviceAccountName: calico-kube-controllers
          containers:
            - name: calico-kube-controllers
              image: quay.io/calico/kube-controllers:v1.0.0
              env:
                # The location of the Calico etcd cluster.
                - name: ETCD_ENDPOINTS
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_endpoints
                # Location of the CA certificate for etcd.
                - name: ETCD_CA_CERT_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_ca
                # Location of the client key for etcd.
                - name: ETCD_KEY_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_key
                # Location of the client certificate for etcd.
                - name: ETCD_CERT_FILE
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: etcd_cert
                # The location of the Kubernetes API. Use the default Kubernetes
                # service for API access.
                - name: K8S_API
                  value: "https://{{.Cluster.Kubernetes.API.Domain}}:443"
                # Since we're running in the host namespace and might not have KubeDNS
                # access, configure the container's /etc/hosts to resolve
                # kubernetes.default to the correct service clusterIP.
                - name: CONFIGURE_ETC_HOSTS
                  value: "true"
              resources:
                requests:
                  cpu: 30m
                  memory: 90Mi
              volumeMounts:
                # Mount in the etcd TLS secrets.
                - mountPath: /etc/kubernetes/ssl/etcd
                  name: etcd-certs
          volumes:
            # Mount in the etcd TLS secrets.
            - name: etcd-certs
              hostPath:
                path: /etc/kubernetes/ssl/etcd
- path: /srv/kubedns-cm.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: kube-dns
      namespace: kube-system
      labels:
        addonmanager.kubernetes.io/mode: EnsureExists
- path: /srv/kubedns-sa.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: kube-dns
      namespace: kube-system
      labels:
        kubernetes.io/cluster-service: "true"
        addonmanager.kubernetes.io/mode: Reconcile
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
        addonmanager.kubernetes.io/mode: Reconcile
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
        spec:
          tolerations:
          - key: "CriticalAddonsOnly"
            operator: "Exists"
          containers:
          - name: kubedns
            image: gcr.io/google_containers/k8s-dns-kube-dns-amd64:1.14.5
            volumeMounts:
            - name: kube-dns-config
              mountPath: /kube-dns-config
            - name: config
              mountPath: /etc/kubernetes/config/
              readOnly: false
            - name: ssl
              mountPath: /etc/kubernetes/ssl/
              readOnly: false
            resources:
              limits:
                memory: 170Mi
              requests:
                cpu: 100m
                memory: 70Mi
            args:
            # command = "/kube-dns
            - --dns-port=10053
            - --domain={{.Cluster.Kubernetes.Domain}}
            - --config-dir=/kube-dns-config
            - --v=2
            - --kubecfg-file=/etc/kubernetes/config/kubelet-kubeconfig.yml
            - --kube-master-url=https://{{.Cluster.Kubernetes.API.Domain}}
            env:
            - name: PROMETHEUS_PORT
              value: "10055"
            ports:
            - containerPort: 10053
              name: dns-local
              protocol: UDP
            - containerPort: 10053
              name: dns-tcp-local
              protocol: TCP
            - containerPort: 10055
              name: metrics
              protocol: TCP
            livenessProbe:
              httpGet:
                path: /healthcheck/kubedns
                port: 10054
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
              initialDelaySeconds: 3
              timeoutSeconds: 5
          - name: dnsmasq
            image: gcr.io/google_containers/k8s-dns-dnsmasq-nanny-amd64:1.14.5
            livenessProbe:
              httpGet:
                path: /healthcheck/dnsmasq
                port: 10054
                scheme: HTTP
              initialDelaySeconds: 60
              timeoutSeconds: 5
              successThreshold: 1
              failureThreshold: 5
            args:
            - -v=2
            - -logtostderr
            - -configDir=/etc/k8s/dns/dnsmasq-nanny
            - -restartDnsmasq=true
            - --
            - -k
            - --cache-size=1000
            - --no-resolv
            - --log-facility=-
            - --server=127.0.0.1#10053
            - --server=/{{.Cluster.Kubernetes.Domain}}/127.0.0.1#10053
            - --server=/in-addr.arpa/127.0.0.1#10053
            - --server=/ip6.arpa/127.0.0.1#10053
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
            volumeMounts:
            - name: kube-dns-config
              mountPath: /etc/k8s/dns/dnsmasq-nanny
          - name: sidecar
            image: gcr.io/google_containers/k8s-dns-sidecar-amd64:1.14.5
            livenessProbe:
              httpGet:
                path: /metrics
                port: 10054
                scheme: HTTP
              initialDelaySeconds: 60
              timeoutSeconds: 5
              successThreshold: 1
              failureThreshold: 5
            args:
            - --v=2
            - --logtostderr
            - --probe=kubedns,127.0.0.1:10053,kubernetes.default.svc.{{.Cluster.Kubernetes.Domain}},5,A
            - --probe=dnsmasq,127.0.0.1:53,kubernetes.default.svc.{{.Cluster.Kubernetes.Domain}},5,A
            ports:
            - containerPort: 10054
              name: metrics
              protocol: TCP
            resources:
              requests:
                memory: 20Mi
                cpu: 10m
          dnsPolicy: Default  # Don't use cluster DNS.
          serviceAccountName: kube-dns
          volumes:
          - name: kube-dns-config
            configMap:
              name: kube-dns
              optional: true
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
        addonmanager.kubernetes.io/mode: Reconcile
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
- path: /srv/default-backend-dep.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: default-http-backend
      namespace: kube-system
      labels:
        k8s-app: default-http-backend
    spec:
      replicas: 2
      template:
        metadata:
          labels:
            k8s-app: default-http-backend
        spec:
          containers:
          - name: default-http-backend
            image: gcr.io/google_containers/defaultbackend:1.0
            livenessProbe:
              httpGet:
                path: /healthz
                port: 8080
                scheme: HTTP
              initialDelaySeconds: 30
              timeoutSeconds: 5
            ports:
            - containerPort: 8080
            resources:
              limits:
                cpu: 10m
                memory: 20Mi
              requests:
                cpu: 10m
                memory: 20Mi
- path: /srv/default-backend-svc.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Service
    metadata:
      name: default-http-backend
      namespace: kube-system
      labels:
        k8s-app: default-http-backend
    spec:
      type: NodePort
      ports:
      - port: 80
        targetPort: 8080
      selector:
        k8s-app: default-http-backend
- path: /srv/ingress-controller-cm.yml
  owner: root
  permissions: 0644
  content: |
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
- path: /srv/ingress-controller-dep.yml
  owner: root
  permissions: 0644
  content: |
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
      replicas: {{len .Cluster.Workers}}
      strategy:
        type: RollingUpdate
        rollingUpdate:
          maxSurge: 1
          maxUnavailable: 1
      template:
        metadata:
          labels:
            k8s-app: nginx-ingress-controller
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ''
        spec:
          affinity:
            podAntiAffinity:
              preferredDuringSchedulingIgnoredDuringExecution:
              - weight: 100
                podAffinityTerm:
                  labelSelector:
                    matchExpressions:
                      - key: k8s-app
                        operator: In
                        values:
                        - nginx-ingress-controller
                  topologyKey: kubernetes.io/hostname
          serviceAccountName: nginx-ingress-controller
          containers:
          - name: nginx-ingress-controller
            image: quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.9.0-beta.17
            args:
            - /nginx-ingress-controller
            - --default-backend-service=$(POD_NAMESPACE)/default-http-backend
            - --configmap=$(POD_NAMESPACE)/ingress-nginx
            - --enable-ssl-passthrough
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
            resources:
              requests:
                memory: "350Mi"
                cpu: "500m"
            livenessProbe:
              httpGet:
                path: /healthz
                port: 10254
                scheme: HTTP
              initialDelaySeconds: 10
              timeoutSeconds: 1
            lifecycle:
              preStop:
                exec:
                  command:
                  - sleep
                  - "15"
            ports:
            - containerPort: 80
              hostPort: 80
            - containerPort: 443
              hostPort: 443
- path: /srv/ingress-controller-svc.yml
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
- path: /srv/kube-proxy-sa.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: kube-proxy
      namespace: kube-system
- path: /srv/kube-proxy-ds.yaml
  owner: root
  permissions: 0644
  content: |
    kind: DaemonSet
    apiVersion: extensions/v1beta1
    metadata:
      name: kube-proxy
      namespace: kube-system
      labels:
        component: kube-proxy
        k8s-app: kube-proxy
        kubernetes.io/cluster-service: "true"
    spec:
      selector:
        matchLabels:
          k8s-app: kube-proxy
      updateStrategy:
        type: RollingUpdate
        rollingUpdate:
          maxUnavailable: 1
      template:
        metadata:
          labels:
            component: kube-proxy
            k8s-app: kube-proxy
            kubernetes.io/cluster-service: "true"
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ''
        spec:
          tolerations:
          - key: node-role.kubernetes.io/master
            operator: Exists
            effect: NoSchedule
          - key: CriticalAddonsOnly
            operator: Exists
          hostNetwork: true
          serviceAccountName: kube-proxy
          containers:
            - name: kube-proxy
              image: quay.io/giantswarm/hyperkube:v1.8.4_coreos.0
              command:
              - /hyperkube
              - proxy
              - --proxy-mode=iptables
              - --logtostderr=true
              - --kubeconfig=/etc/kubernetes/config/proxy-kubeconfig.yml
              - --conntrack-max-per-core=131072
              - --v=2
              livenessProbe:
                httpGet:
                  path: /healthz
                  port: 10256
                initialDelaySeconds: 10
                periodSeconds: 3
              resources:
                requests:
                  memory: "80Mi"
                  cpu: "75m"
              securityContext:
                privileged: true
              volumeMounts:
              - mountPath: /etc/ssl/certs
                name: ssl-certs-host
                readOnly: true
              - mountPath: /etc/kubernetes/config/
                name: config-kubernetes
                readOnly: true
              - mountPath: /etc/kubernetes/ssl
                name: ssl-certs-kubernetes
                readOnly: true
          volumes:
          - hostPath:
              path: /etc/kubernetes/config/
            name: config-kubernetes
          - hostPath:
              path: /etc/kubernetes/ssl
            name: ssl-certs-kubernetes
          - hostPath:
              path: /usr/share/ca-certificates
            name: ssl-certs-host

- path: /srv/rbac_bindings.yaml
  owner: root
  permissions: 0644
  content: |
    ## User
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: giantswarm-admin
    subjects:
    - kind: User
      name: {{.Cluster.Kubernetes.API.Domain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: cluster-admin
      apiGroup: rbac.authorization.k8s.io
    ---
    ## Worker
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: kubelet
    subjects:
    - kind: User
      name: {{.Cluster.Kubernetes.Kubelet.Domain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: system:node
      apiGroup: rbac.authorization.k8s.io
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: proxy
    subjects:
    - kind: User
      name: {{.Cluster.Kubernetes.Kubelet.Domain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: system:node-proxier
      apiGroup: rbac.authorization.k8s.io
    ---
    ## Master
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: kube-controller-manager
    subjects:
    - kind: User
      name: {{.Cluster.Kubernetes.API.Domain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: system:kube-controller-manager
      apiGroup: rbac.authorization.k8s.io
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: kube-scheduler
    subjects:
    - kind: User
      name: {{.Cluster.Kubernetes.API.Domain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: system:kube-scheduler
      apiGroup: rbac.authorization.k8s.io
    ---
    ## Calico
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: calico-kube-controllers
    subjects:
    - kind: ServiceAccount
      name: calico-kube-controllers
      namespace: kube-system
    roleRef:
      kind: ClusterRole
      name: calico-kube-controllers
      apiGroup: rbac.authorization.k8s.io
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: calico-node
    subjects:
    - kind: ServiceAccount
      name: calico-node
      namespace: kube-system
    roleRef:
      kind: ClusterRole
      name: calico-node
      apiGroup: rbac.authorization.k8s.io
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: calico-node-controller
    subjects:
    - kind: ServiceAccount
      name: calico-node-controller
      namespace: kube-system
    roleRef:
      kind: ClusterRole
      name: calico-node-controller
      apiGroup: rbac.authorization.k8s.io
    ---
    ## DNS
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: kube-dns
    subjects:
    - kind: ServiceAccount
      name: kube-dns
      namespace: kube-system
    roleRef:
      kind: ClusterRole
      name: system:kube-dns
      apiGroup: rbac.authorization.k8s.io
    ---
    ## IC
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: nginx-ingress-controller
    subjects:
    - kind: ServiceAccount
      name: nginx-ingress-controller
      namespace: kube-system
    roleRef:
      kind: ClusterRole
      name: nginx-ingress-controller
      apiGroup: rbac.authorization.k8s.io
    ---
    kind: RoleBinding
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: nginx-ingress-controller
      namespace: kube-system
    subjects:
    - kind: ServiceAccount
      name: nginx-ingress-controller
      namespace: kube-system
    roleRef:
      kind: Role
      name: nginx-ingress-role
      apiGroup: rbac.authorization.k8s.io
- path: /srv/rbac_roles.yaml
  owner: root
  permissions: 0644
  content: |
    ## Calico
    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
    rules:
      - apiGroups:
        - ""
        - extensions
        resources:
          - pods
          - namespaces
          - networkpolicies
        verbs:
          - watch
          - list
    ---
    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: calico-node
      namespace: kube-system
    rules:
      - apiGroups: [""]
        resources:
          - pods
          - nodes
        verbs:
          - get
    ---
    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1beta1
    metadata:
      name: calico-node-controller
      namespace: kube-system
    rules:
      - apiGroups:
        - ""
        - extensions
        resources:
          - nodes
        verbs:
          - watch
          - list
    ---
    ## IC
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: nginx-ingress-controller
      namespace: kube-system
    ---
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRole
    metadata:
      name: nginx-ingress-controller
      namespace: kube-system
    rules:
      - apiGroups:
          - ""
        resources:
          - configmaps
          - endpoints
          - nodes
          - pods
          - secrets
        verbs:
          - list
          - watch
      - apiGroups:
          - ""
        resources:
          - nodes
        verbs:
          - get
      - apiGroups:
          - ""
        resources:
          - services
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - "extensions"
        resources:
          - ingresses
        verbs:
          - get
          - list
          - watch
      - apiGroups:
          - ""
        resources:
            - events
        verbs:
            - create
            - patch
      - apiGroups:
          - "extensions"
        resources:
          - ingresses/status
        verbs:
          - update
    ---
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: Role
    metadata:
      name: nginx-ingress-role
      namespace: kube-system
    rules:
      - apiGroups:
          - ""
        resources:
          - configmaps
          - pods
          - secrets
          - namespaces
        verbs:
          - get
      - apiGroups:
          - ""
        resources:
          - configmaps
        resourceNames:
          # Defaults to "<election-id>-<ingress-class>"
          # Here: "<ingress-controller-leader>-<nginx>"
          # This has to be adapted if you change either parameter
          # when launching the nginx-ingress-controller.
          - "ingress-controller-leader-nginx"
        verbs:
          - get
          - update
      - apiGroups:
          - ""
        resources:
          - configmaps
        verbs:
          - create
      - apiGroups:
          - ""
        resources:
          - endpoints
        verbs:
          - get
          - create
          - update
- path: /srv/psp_policies.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: extensions/v1beta1
    kind: PodSecurityPolicy
    metadata:
      name: privileged
    spec:
      allowPrivilegeEscalation: true
      fsGroup:
        rule: RunAsAny
      privileged: true
      runAsUser:
        rule: RunAsAny
      seLinux:
        rule: RunAsAny
      supplementalGroups:
        rule: RunAsAny
      volumes:
      - '*'
      hostPID: true
      hostIPC: true
      hostNetwork: true
      hostPorts:
      - min: 1
        max: 65536
    ---
    apiVersion: extensions/v1beta1
    kind: PodSecurityPolicy
    metadata:
      name: restricted
    spec:
      privileged: false
      fsGroup:
        rule: RunAsAny
      runAsUser:
        rule: RunAsAny
      seLinux:
        rule: RunAsAny
      supplementalGroups:
        rule: RunAsAny
      volumes:
      - 'emptyDir'
      - 'secret'
      - 'downwardAPI'
      - 'configMap'
      - 'persistentVolumeClaim'
      - 'projected'
      hostPID: false
      hostIPC: false
      hostNetwork: false
- path: /srv/psp_roles.yaml
  owner: root
  permissions: 0644
  content: |
    # restrictedPSP grants access to use
    # the restricted PSP.
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRole
    metadata:
      name: restricted-psp-user
    rules:
    - apiGroups:
      - extensions
      resources:
      - podsecuritypolicies
      resourceNames:
      - restricted
      verbs:
      - use
    ---
    # privilegedPSP grants access to use the privileged
    # PSP.
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRole
    metadata:
      name: privileged-psp-user
    rules:
    - apiGroups:
      - extensions
      resources:
      - podsecuritypolicies
      resourceNames:
      - privileged
      verbs:
      - use
- path: /srv/psp_binding.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRoleBinding
    metadata:
        name: privileged-psp-users
    subjects:
    - kind: ServiceAccount
      name: calico-node
      namespace: kube-system
    - kind: ServiceAccount
      name: calico-kube-controllers
      namespace: kube-system
    - kind: ServiceAccount
      name: calico-node-controller
      namespace: kube-system
    - kind: ServiceAccount
      name: calico-ipip-pinger
      namespace: kube-system
    - kind: ServiceAccount
      name: kube-dns
      namespace: kube-system
    - kind: ServiceAccount
      name: kube-proxy
      namespace: kube-system
    - kind: ServiceAccount
      name: nginx-ingress-controller
      namespace: kube-system
    roleRef:
       apiGroup: rbac.authorization.k8s.io
       kind: ClusterRole
       name: privileged-psp-user
    ---
    # grants the restricted PSP role to
    # the all authenticated users.
    apiVersion: rbac.authorization.k8s.io/v1beta1
    kind: ClusterRoleBinding
    metadata:
        name: restricted-psp-users
    subjects:
    - kind: Group
      apiGroup: rbac.authorization.k8s.io
      name: system:authenticated
    roleRef:
       apiGroup: rbac.authorization.k8s.io
       kind: ClusterRole
       name: restricted-psp-user
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
- path: /opt/k8s-addons
  permissions: 0544
  content: |
      #!/bin/bash

      export KUBECONFIG=/etc/kubernetes/config/addons-kubeconfig.yml
      # kubectl 1.8.4
      KUBECTL=quay.io/giantswarm/docker-kubectl:8cabd75bacbcdad7ac5d85efc3ca90c2fabf023b

      /usr/bin/docker pull $KUBECTL

      # wait for healthy master
      while [ "$(/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL get cs | grep Healthy | wc -l)" -ne "3" ]; do sleep 1 && echo 'Waiting for healthy k8s'; done

      # apply Security bootstrap (RBAC and PSP)
      SECURITY_FILES="rbac_bindings.yaml rbac_roles.yaml psp_policies.yaml psp_roles.yaml psp_binding.yaml"

      for manifest in $SECURITY_FILES
      do
          while
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/$manifest
              [ "$?" -ne "0" ]
          do
              echo "failed to apply /src/$manifest, retrying in 5 sec"
              sleep 5s
          done
      done

      # apply calico CNI
      CALICO_FILES="calico-configmap.yaml\
       calico-node-sa.yaml\
       calico-kube-controllers-sa.yaml\
       calico-ds.yaml\
       calico-kube-controllers.yaml\
       calico-node-controller-sa.yaml\
       calico-node-controller.yaml\
       calico-ipip-pinger-sa.yaml\
       calico-ipip-pinger.yaml"

      for manifest in $CALICO_FILES
      do
          while
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/$manifest
              [ "$?" -ne "0" ]
          do
              echo "failed to apply /src/$manifest, retrying in 5 sec"
              sleep 5s
          done
      done

      # wait for healthy calico - we check for pods - desired vs ready
      while
          # result of this is 'eval [ "$DESIRED_POD_COUNT" -eq "$READY_POD_COUNT" ]'
          /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system  get ds calico-node 2>/dev/null >/dev/null
          RET_CODE_1=$?
          eval $(/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL -n kube-system get ds calico-node | tail -1 | awk '{print "[ \"" $2"\" -eq \""$4"\" ] "}')
          RET_CODE_2=$?
          [ "$RET_CODE_1" -ne "0" ] || [ "$RET_CODE_2" -ne "0" ]
      do
          echo "Waiting for calico to be ready . . "
          sleep 3s
      done

      # apply default storage class
      if [ -f /srv/default-storage-class.yaml ]; then
          while
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/default-storage-class.yaml
              [ "$?" -ne "0" ]
          do
              echo "failed to apply /srv/default-storage-class.yaml, retrying in 5 sec"
              sleep 5s
          done
      else
          echo "no default storage class to apply"
      fi

      # apply k8s addons
      MANIFESTS="kube-proxy-sa.yaml\
                 kube-proxy-ds.yaml\
                 kubedns-cm.yaml\
                 kubedns-sa.yaml\
                 kubedns-dep.yaml\
                 kubedns-svc.yaml\
                 default-backend-dep.yml\
                 default-backend-svc.yml\
                 ingress-controller-cm.yml\
                 ingress-controller-dep.yml\
                 ingress-controller-svc.yml"

      for manifest in $MANIFESTS
      do
          while
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/$manifest
              [ "$?" -ne "0" ]
          do
              echo "failed to apply /srv/$manifest, retrying in 5 sec"
              sleep 5s
          done
      done
      echo "Addons successfully installed"
- path: /etc/kubernetes/config/addons-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: proxy
      user:
        client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
        server: https://{{.Cluster.Kubernetes.API.Domain}}
    contexts:
    - context:
        cluster: local
        user: proxy
      name: service-account-context
    current-context: service-account-context

- path: /etc/kubernetes/config/proxy-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: proxy
      user:
        client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
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
        client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
        server: https://{{.Cluster.Kubernetes.API.Domain}}
    contexts:
    - context:
        cluster: local
        user: kubelet
      name: service-account-context
    current-context: service-account-context
- path: /etc/kubernetes/config/controller-manager-kubeconfig.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Config
    users:
    - name: controller-manager
      user:
        client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
        server: https://{{.Cluster.Kubernetes.API.Domain}}
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
        client-certificate: /etc/kubernetes/ssl/apiserver-crt.pem
        client-key: /etc/kubernetes/ssl/apiserver-key.pem
    clusters:
    - name: local
      cluster:
        certificate-authority: /etc/kubernetes/ssl/apiserver-ca.pem
        server: https://{{.Cluster.Kubernetes.API.Domain}}
    contexts:
    - context:
        cluster: local
        user: scheduler
      name: service-account-context
    current-context: service-account-context

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
  - name: set-ownership-etcd-data-dir.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Set ownership to etcd3 data dir
      Wants=network-online.target

      [Service]
      Type=oneshot
      RemainAfterExit=yes
      TimeoutStartSec=0
      ExecStartPre=/bin/bash -c "/usr/bin/mkdir -p /etc/kubernetes/data/etcd; /usr/bin/chown etcd:etcd /etc/kubernetes/data/etcd"
      ExecStart=/usr/bin/chmod -R 700 /etc/kubernetes/data/etcd
  - name: docker.service
    enable: true
    command: start
    drop-ins:
    - name: 10-giantswarm-extra-args.conf
      content: |
        [Service]
        Environment="DOCKER_CGROUPS=--exec-opt native.cgroupdriver=cgroupfs {{.Cluster.Docker.Daemon.ExtraArgs}}"
        Environment="DOCKER_OPT_BIP=--bip={{.Cluster.Docker.Daemon.CIDR}}"
        Environment="DOCKER_OPTS=--live-restore --icc=false --disable-legacy-registry=true --userland-proxy=false"
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
  - name: etcd2.service
    command: stop
    enable: false
    mask: true
  - name: etcd3.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=etcd3
      Requires=k8s-setup-network-env.service
      After=k8s-setup-network-env.service
      Conflicts=etcd.service etcd2.service

      [Service]
      StartLimitIntervalSec=0
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      LimitNOFILE=40000
      Environment=IMAGE=quay.io/coreos/etcd:v3.2.7
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
          -v /etc/kubernetes/data/etcd/:/var/lib/etcd  \
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
          --advertise-client-urls=https://{{ .Cluster.Etcd.Domain }}:443 \
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
    enable: true
    command: start
    content: |
      [Unit]
      Description=etcd defragmentation job
      After=docker.service etcd3.service
      Requires=docker.service etcd3.service

      [Service]
      Type=oneshot
      EnvironmentFile=/etc/network-environment
      Environment=IMAGE=quay.io/coreos/etcd:v3.2.7
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
    enable: true
    command: start
    content: |
      [Unit]
      Description=Execute etcd3-defragmentation every day at 3.30AM UTC

      [Timer]
      OnCalendar=*-*-* 03:30:00 UTC

      [Install]
      WantedBy=multi-user.target
  - name: k8s-kubelet.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-kubelet
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE=quay.io/giantswarm/hyperkube:v1.8.4_coreos.0"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/bin/sh -c "/usr/bin/docker run --rm --pid=host --net=host --privileged=true \
      -v /:/rootfs:ro,shared \
      -v /sys:/sys:ro \
      -v /dev:/dev:rw \
      -v /var/log:/var/log:rw \
      -v /run/calico/:/run/calico/:rw \
      -v /run/docker/:/run/docker/:rw \
      -v /run/docker.sock:/run/docker.sock:rw \
      -v /usr/lib/os-release:/etc/os-release \
      -v /usr/share/ca-certificates/:/etc/ssl/certs \
      -v /var/lib/docker/:/var/lib/docker:rw,shared \
      -v /var/lib/kubelet/:/var/lib/kubelet:rw,shared \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      -v /etc/cni/net.d/:/etc/cni/net.d/ \
      -v /opt/cni/bin/:/opt/cni/bin/ \
      -v /usr/sbin/iscsiadm:/usr/sbin/iscsiadm \
      -v /etc/iscsi/:/etc/iscsi/ \
      -v /dev/disk/by-path/:/dev/disk/by-path/ \
      -v /dev/mapper/:/dev/mapper/ \
      -v /usr/sbin/mkfs.xfs:/usr/sbin/mkfs.xfs \
      -v /usr/lib64/libxfs.so.0:/usr/lib/libxfs.so.0 \
      -v /usr/lib64/libxcmd.so.0:/usr/lib/libxcmd.so.0 \
      -e ETCD_CA_CERT_FILE=/etc/kubernetes/ssl/etcd/server-ca.pem \
      -e ETCD_CERT_FILE=/etc/kubernetes/ssl/etcd/server-crt.pem \
      -e ETCD_KEY_FILE=/etc/kubernetes/ssl/etcd/server-key.pem \
      --name $NAME \
      $IMAGE \
      /hyperkube kubelet \
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
      --register-with-taints=node-role.kubernetes.io/master=:NoSchedule \
      --allow-privileged=true \
      --kubeconfig=/etc/kubernetes/config/kubelet-kubeconfig.yml \
      --node-labels="node-role.kubernetes.io/master,role=master,kubernetes.io/hostname=${HOSTNAME},ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
      --kube-reserved="cpu=150m,memory=250Mi" \
      --system-reserved="cpu=150m,memory=250Mi" \
      --v=2"
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: update-engine.service
    enable: false
    command: stop
    mask: true
  - name: locksmithd.service
    enable: false
    command: stop
    mask: true
  - name: fleet.service
    enable: false
    mask: true
    command: stop
  - name: fleet.socket
    enable: false
    mask: true
    command: stop
  - name: flanneld.service
    enable: false
    command: stop
    mask: true
  - name: systemd-networkd-wait-online.service
    enable: true
    command: start
  - name: k8s-api-server.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-api-server
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE=quay.io/giantswarm/hyperkube:v1.8.4_coreos.0"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/mkdir -p /etc/kubernetes/manifests
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --name $NAME --net=host \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/secrets/token_sign_key.pem:/etc/kubernetes/secrets/token_sign_key.pem \
      -v /etc/kubernetes/encryption/:/etc/kubernetes/encryption \
      $IMAGE \
      /hyperkube apiserver \
      --allow_privileged=true \
      --insecure_bind_address=0.0.0.0 \
      --anonymous-auth=false \
      --insecure-port=0 \
      --kubelet_https=true \
      --kubelet-preferred-address-types=InternalIP \
      --secure_port={{.Cluster.Kubernetes.API.SecurePort}} \
      --bind-address=${DEFAULT_IPV4} \
      --etcd-prefix={{.Cluster.Etcd.Prefix}} \
      --profiling=false \
      --repair-malformed-updates=false \
      --service-account-lookup=true \
      --authorization-mode=RBAC \
      --admission-control=NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota,DefaultStorageClass,PodSecurityPolicy \
      --cloud-provider={{.Cluster.Kubernetes.CloudProvider}} \
      --service-cluster-ip-range={{.Cluster.Kubernetes.API.ClusterIPRange}} \
      --etcd-servers=https://127.0.0.1:2379 \
      --etcd-cafile=/etc/kubernetes/ssl/etcd/server-ca.pem \
      --etcd-certfile=/etc/kubernetes/ssl/etcd/server-crt.pem \
      --etcd-keyfile=/etc/kubernetes/ssl/etcd/server-key.pem \
      --advertise-address=${DEFAULT_IPV4} \
      --runtime-config=api/all=true \
      --logtostderr=true \
      --tls-cert-file=/etc/kubernetes/ssl/apiserver-crt.pem \
      --tls-private-key-file=/etc/kubernetes/ssl/apiserver-key.pem \
      --client-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem \
      --service-account-key-file=/etc/kubernetes/ssl/service-account-key.pem \
      --audit-log-path=/var/log/apiserver/audit.log \
      --audit-log-maxage=30 \
      --audit-log-maxbackup=10 \
      --audit-log-maxsize=100 \
      --experimental-encryption-provider-config=/etc/kubernetes/encryption/k8s-encryption-config.yaml
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-controller-manager.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-controller-manager Service
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE=quay.io/giantswarm/hyperkube:v1.8.4_coreos.0"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host --name $NAME \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      -v /etc/kubernetes/secrets/token_sign_key.pem:/etc/kubernetes/secrets/token_sign_key.pem \
      $IMAGE \
      /hyperkube controller-manager \
      --logtostderr=true \
      --v=2 \
      --cloud-provider={{.Cluster.Kubernetes.CloudProvider}} \
      --profiling=false \
      --terminated-pod-gc-threshold=10 \
      --use-service-account-credentials=true \
      --kubeconfig=/etc/kubernetes/config/controller-manager-kubeconfig.yml \
      --root-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem \
      --service-account-private-key-file=/etc/kubernetes/ssl/service-account-key.pem
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-scheduler.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-scheduler Service
      StartLimitIntervalSec=0

      [Service]
      Restart=always
      RestartSec=0
      TimeoutStopSec=10
      EnvironmentFile=/etc/network-environment
      Environment="IMAGE=quay.io/giantswarm/hyperkube:v1.8.4_coreos.0"
      Environment="NAME=%p.service"
      Environment="NETWORK_CONFIG_CONTAINER="
      ExecStartPre=/usr/bin/docker pull $IMAGE
      ExecStartPre=-/usr/bin/docker stop -t 10 $NAME
      ExecStartPre=-/usr/bin/docker rm -f $NAME
      ExecStart=/usr/bin/docker run --rm --net=host --name $NAME \
      -v /etc/kubernetes/ssl/:/etc/kubernetes/ssl/ \
      -v /etc/kubernetes/config/:/etc/kubernetes/config/ \
      $IMAGE \
      /hyperkube scheduler \
      --logtostderr=true \
      --v=2 \
      --profiling=false \
      --kubeconfig=/etc/kubernetes/config/scheduler-kubeconfig.yml
      ExecStop=-/usr/bin/docker stop -t 10 $NAME
      ExecStopPost=-/usr/bin/docker rm -f $NAME
  - name: k8s-addons.service
    enable: true
    command: start
    content: |
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

  update:
    reboot-strategy: off

{{ range .Extension.VerbatimSections }}
{{ .Content }}
{{ end }}
`

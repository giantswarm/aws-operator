package v_3_7_4

const MasterTemplate = `#cloud-config
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
{{ if not .DisableCalico -}}
- path: /srv/calico-all.yaml
  owner: root
  permissions: 644
  content: |
    # CALICO HAS SEPARATE MANIFEST FOR AZURE
    # the azure manifest can be found in: https://github.com/giantswarm/azure-operator/blob/master/service/controller/vX/cloudconfig/template.go
    # where X is the version of azure operator
    #
    # Extra changes:
    #  - Added resource limits to calico-node and calico-kube-controllers.
    #  - Added resource limits to install-cni.
    #  - Added 'priorityClassName: system-cluster-critical' to calico daemonset.
    #
    # Calico Version v3.2.3
    # https://docs.projectcalico.org/v3.2/releases#v3.2.3
    # This manifest includes the following component versions:
    #   calico/node:v3.2.3
    #   calico/cni:v3.2.3
    #   calico/kube-controllers:v3.2.3

    # This ConfigMap is used to configure a self-hosted Calico installation.
    kind: ConfigMap
    apiVersion: v1
    metadata:
      name: calico-config
      namespace: kube-system
    data:
      # Configure this with the location of your etcd cluster.
      etcd_endpoints: "https://{{ .Cluster.Etcd.Domain }}:{{ .EtcdPort }}"

      # If you're using TLS enabled etcd uncomment the following.
      # You must also populate the Secret below with these files.
      etcd_ca: "/calico-secrets/client-ca.pem"
      etcd_cert: "/calico-secrets/client-crt.pem"
      etcd_key: "/calico-secrets/client-key.pem"
      # Configure the Calico backend to use.
      calico_backend: "bird"

      # Configure the MTU to use
      veth_mtu: "{{.Cluster.Calico.MTU}}"

      # The CNI network configuration to install on each node.  The special
      # values in this config will be automatically populated.
      cni_network_config: |-
        {
          "name": "k8s-pod-network",
          "cniVersion": "0.3.0",
          "plugins": [
            {
              "type": "calico",
              "log_level": "info",
              "etcd_endpoints": "__ETCD_ENDPOINTS__",
              "etcd_key_file": "__ETCD_KEY_FILE__",
              "etcd_cert_file": "__ETCD_CERT_FILE__",
              "etcd_ca_cert_file": "__ETCD_CA_CERT_FILE__",
              "mtu": __CNI_MTU__,
              "ipam": {
                  "type": "calico-ipam"
              },
              "policy": {
                  "type": "k8s"
              },
              "kubernetes": {
                  "kubeconfig": "__KUBECONFIG_FILEPATH__"
              }
            },
            {
              "type": "portmap",
              "snat": true,
              "capabilities": {"portMappings": true}
            }
          ]
        }

    ---


    # The following contains k8s Secrets for use with a TLS enabled etcd cluster.
    # For information on populating Secrets, see http://kubernetes.io/docs/user-guide/secrets/
    apiVersion: v1
    kind: Secret
    type: Opaque
    metadata:
      name: calico-etcd-secrets
      namespace: kube-system
    data:
      # Populate the following files with etcd TLS configuration if desired, but leave blank if
      # not using TLS for etcd.
      # This self-hosted install expects three files with the following names.  The values
      # should be base64 encoded strings of the entire contents of each file.
      # etcd-key: null
      # etcd-cert: null
      # etcd-ca: null

    ---

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
            # This, along with the CriticalAddonsOnly toleration below,
            # marks the pod as a critical add-on, ensuring it gets
            # priority scheduling and that its resources are reserved
            # if it ever gets evicted.
            scheduler.alpha.kubernetes.io/critical-pod: ''
        spec:
          nodeSelector:
            beta.kubernetes.io/os: linux
          hostNetwork: true
          tolerations:
            # Make sure calico-node gets scheduled on all nodes.
            - effect: NoSchedule
              operator: Exists
            # Mark the pod as a critical add-on for rescheduling.
            - key: CriticalAddonsOnly
              operator: Exists
            - effect: NoExecute
              operator: Exists
          serviceAccountName: calico-node
          priorityClassName: system-cluster-critical
          # Minimize downtime during a rolling upgrade or deletion; tell Kubernetes to do a "force
          # deletion": https://kubernetes.io/docs/concepts/workloads/pods/pod/#termination-of-pods.
          terminationGracePeriodSeconds: 0
          containers:
            # Runs calico/node container on each Kubernetes node.  This
            # container programs network policy and routes on each
            # host.
            - name: calico-node
              image: {{ .RegistryDomain }}/giantswarm/node:v3.2.3
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
                # Set noderef for node controller.
                - name: CALICO_K8S_NODE_REF
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
                # Choose the backend to use.
                - name: CALICO_NETWORKING_BACKEND
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: calico_backend
                # Cluster type to identify the deployment type
                - name: CLUSTER_TYPE
                  value: "k8s,bgp"
                # Auto-detect the BGP IP address.
                - name: IP
                  value: "autodetect"
                # Enable IPIP
                - name: CALICO_IPV4POOL_IPIP
                  value: "Always"
                # Set MTU for tunnel device used if ipip is enabled
                - name: FELIX_IPINIPMTU
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: veth_mtu
                # The default IPv4 pool to create on startup if none exists. Pod IPs will be
                # chosen from this range. Changing this value after installation will have
                # no effect. This should fall within --cluster-cidr.
                - name: CALICO_IPV4POOL_CIDR
                  value: "{{.Cluster.Calico.Subnet}}/{{.Cluster.Calico.CIDR}}"
                # Disable file logging so kubectl logs works.
                - name: CALICO_DISABLE_FILE_LOGGING
                  value: "true"
                # Set Felix endpoint to host default action to ACCEPT.
                - name: FELIX_DEFAULTENDPOINTTOHOSTACTION
                  value: "ACCEPT"
                # Disable IPv6 on Kubernetes.
                - name: FELIX_IPV6SUPPORT
                  value: "false"
                # Set Felix logging to "info"
                - name: FELIX_LOGSEVERITYSCREEN
                  value: "info"
                - name: FELIX_HEALTHENABLED
                  value: "true"
              securityContext:
                privileged: true
              resources:
                requests:
                  cpu: 250m
                  memory: 150Mi
                limits:
                  cpu: 250m
                  memory: 150Mi
              livenessProbe:
                httpGet:
                  path: /liveness
                  port: 9099
                  host: localhost
                periodSeconds: 10
                initialDelaySeconds: 10
                failureThreshold: 6
              readinessProbe:
                exec:
                  command:
                  - /bin/calico-node
                  - -bird-ready
                  - -felix-ready
                periodSeconds: 10
              volumeMounts:
                - mountPath: /lib/modules
                  name: lib-modules
                  readOnly: true
                - mountPath: /var/run/calico
                  name: var-run-calico
                  readOnly: false
                - mountPath: /var/lib/calico
                  name: var-lib-calico
                  readOnly: false
                - mountPath: /calico-secrets
                  name: etcd-certs
            # This container installs the Calico CNI binaries
            # and CNI network config file on each node.
            - name: install-cni
              image: {{ .RegistryDomain }}/giantswarm/cni:v3.2.3
              command: ["/install-cni.sh"]
              env:
                # Name of the CNI config file to create.
                - name: CNI_CONF_NAME
                  value: "10-calico.conflist"
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
                # CNI MTU Config variable
                - name: CNI_MTU
                  valueFrom:
                    configMapKeyRef:
                      name: calico-config
                      key: veth_mtu
              # install-cni also monitors etcd certificates,
              # so use reasonable resource limits.
              resources:
                requests:
                  cpu: 50m
                  memory: 100Mi
                limits:
                  cpu: 50m
                  memory: 100Mi
              volumeMounts:
                - mountPath: /host/opt/cni/bin
                  name: cni-bin-dir
                - mountPath: /host/etc/cni/net.d
                  name: cni-net-dir
                - mountPath: /calico-secrets
                  name: etcd-certs
          volumes:
            # Used by calico/node.
            - name: lib-modules
              hostPath:
                path: /lib/modules
            - name: var-run-calico
              hostPath:
                path: /var/run/calico
            - name: var-lib-calico
              hostPath:
                path: /var/lib/calico
            # Used to install CNI.
            - name: cni-bin-dir
              hostPath:
                path: /opt/cni/bin
            - name: cni-net-dir
              hostPath:
                path: /etc/cni/net.d
            # Mount in the etcd TLS secrets.
            - name: etcd-certs
              hostPath:
                path: /etc/kubernetes/ssl/etcd
    ---

    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: calico-node
      namespace: kube-system

    ---

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
      # The controllers can only have a single active instance.
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
          # The controllers must run in the host network namespace so that
          # it isn't governed by policy that would prevent it from working.
          hostNetwork: true
          tolerations:
            # Mark the pod as a critical add-on for rescheduling.
            - key: CriticalAddonsOnly
              operator: Exists
            - key: node-role.kubernetes.io/master
              effect: NoSchedule
          serviceAccountName: calico-kube-controllers
          containers:
            - name: calico-kube-controllers
              image: {{ .RegistryDomain }}/giantswarm/kube-controllers:v3.2.3
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
                # Choose which controllers to run.
                - name: ENABLED_CONTROLLERS
                  value: policy,profile,workloadendpoint,node
              volumeMounts:
                # Mount in the etcd TLS secrets.
                - mountPath: /calico-secrets
                  name: etcd-certs
              resources:
                requests:
                  cpu: 250m
                  memory: 100Mi
                limits:
                  cpu: 250m
                  memory: 100Mi
              readinessProbe:
                exec:
                  command:
                  - /usr/bin/check-status
                  - -r
          volumes:
            # Mount in the etcd TLS secrets.
            - name: etcd-certs
              hostPath:
                path: /etc/kubernetes/ssl/etcd

    ---

    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: calico-kube-controllers
      namespace: kube-system
    ---

    # Calico Version v3.2.3
    # https://docs.projectcalico.org/v3.2/releases#v3.2.3

    ---

    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: calico-kube-controllers
    rules:
      - apiGroups:
        - ""
        - extensions
        resources:
          - pods
          - namespaces
          - networkpolicies
          - nodes
          - serviceaccounts
        verbs:
          - watch
          - list
      - apiGroups:
        - networking.k8s.io
        resources:
          - networkpolicies
        verbs:
          - watch
          - list
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: calico-kube-controllers
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: calico-kube-controllers
    subjects:
    - kind: ServiceAccount
      name: calico-kube-controllers
      namespace: kube-system

    ---

    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: calico-node
    rules:
      - apiGroups: [""]
        resources:
          - pods
          - nodes
        verbs:
          - get

    ---

    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: calico-node
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: calico-node
    subjects:
    - kind: ServiceAccount
      name: calico-node
      namespace: kube-system
{{ end -}}
{{- if not .DisableCoreDNS }}
- path: /srv/coredns.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: coredns
      namespace: kube-system
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      labels:
        kubernetes.io/bootstrapping: rbac-defaults
      name: system:coredns
    rules:
    - apiGroups:
      - ""
      resources:
      - endpoints
      - services
      - pods
      - namespaces
      verbs:
      - list
      - watch
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      annotations:
        rbac.authorization.kubernetes.io/autoupdate: "true"
      labels:
        kubernetes.io/bootstrapping: rbac-defaults
      name: system:coredns
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: system:coredns
    subjects:
    - kind: ServiceAccount
      name: coredns
      namespace: kube-system
    ---
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: coredns
      namespace: kube-system
    data:
      Corefile: |
        .:53 {
            errors
            health
            kubernetes {{.Cluster.Kubernetes.Domain}} {{.Cluster.Kubernetes.API.ClusterIPRange}} {{.Cluster.Calico.Subnet}}/{{.Cluster.Calico.CIDR}} {
              pods insecure
              upstream
              fallthrough in-addr.arpa ip6.arpa
            }
            prometheus :9153
            proxy . /etc/resolv.conf
            cache 30
        }
    ---
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
      name: coredns
      namespace: kube-system
      labels:
        k8s-app: coredns
        kubernetes.io/name: "CoreDNS"
    spec:
      replicas: 2
      strategy:
        type: RollingUpdate
        rollingUpdate:
          maxUnavailable: 1
      selector:
        matchLabels:
          k8s-app: coredns
      template:
        metadata:
          labels:
            k8s-app: coredns
          annotations:
            scheduler.alpha.kubernetes.io/critical-pod: ''
        spec:
          serviceAccountName: coredns
          priorityClassName: system-cluster-critical
          tolerations:
            - key: node-role.kubernetes.io/master
              effect: NoSchedule
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
                        - coredns
                  topologyKey: kubernetes.io/hostname
          containers:
          - name: coredns
            image: {{ .RegistryDomain }}/giantswarm/coredns:1.1.1
            imagePullPolicy: IfNotPresent
            args: [ "-conf", "/etc/coredns/Corefile" ]
            volumeMounts:
            - name: config-volume
              mountPath: /etc/coredns
            ports:
            - containerPort: 53
              name: dns
              protocol: UDP
            - containerPort: 53
              name: dns-tcp
              protocol: TCP
            livenessProbe:
              httpGet:
                path: /health
                port: 8080
                scheme: HTTP
              initialDelaySeconds: 60
              timeoutSeconds: 5
              successThreshold: 1
              failureThreshold: 5
          dnsPolicy: Default
          volumes:
            - name: config-volume
              configMap:
                name: coredns
                items:
                - key: Corefile
                  path: Corefile
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: coredns
      namespace: kube-system
      labels:
        k8s-app: coredns
        kubernetes.io/cluster-service: "true"
        kubernetes.io/name: "CoreDNS"
    spec:
      selector:
        k8s-app: coredns
      clusterIP: {{.Cluster.Kubernetes.DNS.IP}}
      ports:
      - name: dns
        port: 53
        protocol: UDP
      - name: dns-tcp
        port: 53
        protocol: TCP
{{- end }}
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
{{- if not .DisableIngressController }}
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
            image: {{ .RegistryDomain }}/giantswarm/defaultbackend:1.0
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
      enable-vts-status: "true"
      server-name-hash-bucket-size: "1024"
      server-name-hash-max-size: "1024"
      server-tokens: "false"
      worker-processes: "4"
      # Disables setting a 'Strict-Transport-Security' header, which can be harmful.
      # See https://github.com/kubernetes/ingress-nginx/issues/549#issuecomment-291894246
      hsts: "false"
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
          priorityClassName: system-cluster-critical
          initContainers:
          - command:
            - sh
            - -c
            - sysctl -w net.core.somaxconn=32768; sysctl -w net.ipv4.ip_local_port_range="1024 65535"
            image: {{ .RegistryDomain }}/giantswarm/alpine:3.7
            imagePullPolicy: IfNotPresent
            name: sysctl
            securityContext:
              privileged: true
          containers:
          - name: nginx-ingress-controller
            image: {{ .RegistryDomain }}/giantswarm/nginx-ingress-controller:0.12.0
            args:
            - /nginx-ingress-controller
            - --default-backend-service=$(POD_NAMESPACE)/default-http-backend
            - --configmap=$(POD_NAMESPACE)/ingress-nginx
            - --annotations-prefix=nginx.ingress.kubernetes.io
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
{{- end }}
{{- if not .DisableIngressControllerService }}
- path: /srv/ingress-controller-svc.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Service
    metadata:
      annotations:
        prometheus.io/port: "10254"
        prometheus.io/scrape: "true"
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
{{- end }}
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
          hostNetwork: true
          priorityClassName: system-node-critical
          serviceAccountName: kube-proxy
          containers:
            - name: kube-proxy
              image: {{ .RegistryDomain }}/{{ .Images.Kubernetes }}
              command:
              - /hyperkube
              - proxy
              - --config=/etc/kubernetes/config/proxy-config.yml
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
              - mountPath: /var/run/dbus/system_bus_socket
                name: dbus
              - mountPath: /etc/kubernetes/ssl
                name: ssl-certs-kubernetes
                readOnly: true
              - mountPath: /lib/modules
                name: lib-modules
                readOnly: true
          volumes:
          - hostPath:
              path: /etc/kubernetes/config/
            name: config-kubernetes
          - hostPath:
              path: /etc/kubernetes/ssl
            name: ssl-certs-kubernetes
          - hostPath:
              path: /var/run/dbus/system_bus_socket
            name: dbus
          - hostPath:
              path: /usr/share/ca-certificates
            name: ssl-certs-host
          - hostPath:
              path: /lib/modules
            name: lib-modules
- path: /srv/rbac_bindings.yaml
  owner: root
  permissions: 0644
  content: |
    ## User
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    ## node-operator
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: node-operator
    subjects:
    - kind: User
      name: node-operator.{{.BaseDomain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: node-operator
      apiGroup: rbac.authorization.k8s.io
    ---
    ## prometheus-external is prometheus from host cluster
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: prometheus-external
    subjects:
    - kind: User
      name: prometheus.{{.BaseDomain}}
      apiGroup: rbac.authorization.k8s.io
    roleRef:
      kind: ClusterRole
      name: prometheus-external
      apiGroup: rbac.authorization.k8s.io
{{- if not .DisableIngressController }}
    ---
    ## IC
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
{{- end }}
- path: /srv/rbac_roles.yaml
  owner: root
  permissions: 0644
  content: |
    ## node-operator
    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: node-operator
    rules:
    - apiGroups: [""]
      resources: ["nodes"]
      verbs: ["patch"]
    - apiGroups: [""]
      resources: ["pods"]
      verbs: ["list", "delete"]
    ---
    ## prometheus-external
    kind: ClusterRole
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
      name: prometheus-external
    rules:
    - apiGroups: [""]
      resources:
      - nodes
      - nodes/proxy
      - services
      - endpoints
      - pods
      verbs: ["get", "list", "watch"]
    - apiGroups:
      - extensions
      resources:
      - ingresses
      verbs: ["get", "list", "watch"]
    - nonResourceURLs: ["/metrics"]
      verbs: ["get"]
{{- if not .DisableIngressController }}
    ---
    ## IC
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: nginx-ingress-controller
      namespace: kube-system
    ---
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
{{- end }}
- path: /srv/psp_policies.yaml
  owner: root
  permissions: 0644
  content: |
    apiVersion: extensions/v1beta1
    kind: PodSecurityPolicy
    metadata:
      name: privileged
      annotations:
        seccomp.security.alpha.kubernetes.io/allowedProfileNames: '*'
    spec:
      allowPrivilegeEscalation: true
      allowedCapabilities:
      - '*'
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
      - min: 0
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
    apiVersion: rbac.authorization.k8s.io/v1
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
      # kubectl 1.12.2
      KUBECTL={{ .RegistryDomain }}/giantswarm/docker-kubectl:f5cae44c480bd797dc770dd5f62d40b74063c0d7

      /usr/bin/docker pull $KUBECTL

      # wait for healthy master
      while [ "$(/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL get cs | grep Healthy | wc -l)" -ne "3" ]; do sleep 1 && echo 'Waiting for healthy k8s'; done

      # apply Security bootstrap (RBAC and PSP)
      SECURITY_FILES=""
      SECURITY_FILES="${SECURITY_FILES} rbac_bindings.yaml"
      SECURITY_FILES="${SECURITY_FILES} rbac_roles.yaml"
      SECURITY_FILES="${SECURITY_FILES} psp_policies.yaml"
      SECURITY_FILES="${SECURITY_FILES} psp_roles.yaml"
      SECURITY_FILES="${SECURITY_FILES} psp_binding.yaml"

      for manifest in $SECURITY_FILES
      do
          while
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/$manifest
              [ "$?" -ne "0" ]
          do
              echo "failed to apply /srv/$manifest, retrying in 5 sec"
              sleep 5s
          done
      done

      # check for other master and remove it
      THIS_MACHINE=$(cat /etc/hostname)
      for master in $(/usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL get nodes --no-headers=true --selector role=master | awk '{print $1}')
      do
          if [ "$master" != "$THIS_MACHINE" ]; then
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /etc/kubernetes:/etc/kubernetes $KUBECTL delete node $master
          fi
      done

      # wait for etcd dns (return code 35 is bad certificate which is good enough here)
      # to avoid issues with flapping dns once it is changed on an upgrade we better check 10 times in a row.
      n=0
      until [ $n -ge 10 ]
      do
        while
            curl "https://{{ .Cluster.Etcd.Domain }}:{{ .EtcdPort }}" -k 2>/dev/null >/dev/null
            RET_CODE=$?
            [ "$RET_CODE" -ne "35" ]
        do
            n=0 # reset because it failed again
            echo "Waiting for etcd to be ready . . "
            sleep 3s
        done
        n=$[$n+1]
      done

      # install kube-proxy
      PROXY_MANIFESTS="kube-proxy-sa.yaml kube-proxy-ds.yaml"
      for manifest in $PROXY_MANIFESTS
      do
          while
              /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/$manifest
              [ "$?" -ne "0" ]
          do
              echo "failed to apply /srv/$manifest, retrying in 5 sec"
              sleep 5s
          done
      done
      echo "kube-proxy successfully installed"

      {{ if not .DisableCalico -}}

      # apply calico
      CALICO_FILE="calico-all.yaml"

      while
          /usr/bin/docker run -e KUBECONFIG=${KUBECONFIG} --net=host --rm -v /srv:/srv -v /etc/kubernetes:/etc/kubernetes $KUBECTL apply -f /srv/$CALICO_FILE
          [ "$?" -ne "0" ]
      do
          echo "failed to apply /srv/$manifest, retrying in 5 sec"
          sleep 5s
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

      {{ end -}}

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
      MANIFESTS=""
      {{ range .ExtraManifests -}}
      MANIFESTS="${MANIFESTS} {{ . }}"
      {{ end -}}
      {{ if not .DisableCoreDNS }}
      MANIFESTS="${MANIFESTS} coredns.yaml"
      {{ end -}}
      {{ if not .DisableIngressController -}}
      MANIFESTS="${MANIFESTS} default-backend-dep.yml"
      MANIFESTS="${MANIFESTS} default-backend-svc.yml"
      MANIFESTS="${MANIFESTS} ingress-controller-cm.yml"
      MANIFESTS="${MANIFESTS} ingress-controller-dep.yml"
      {{ end -}}
      {{ if not .DisableIngressControllerService -}}
      MANIFESTS="${MANIFESTS} ingress-controller-svc.yml"
      {{ end -}}

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
- path: /etc/kubernetes/config/proxy-config.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: kubeproxy.config.k8s.io/v1alpha1
    clientConnection:
      kubeconfig: /etc/kubernetes/config/proxy-kubeconfig.yml
    kind: KubeProxyConfiguration
    mode: iptables
    resourceContainer: /kube-proxy
    clusterCIDR: {{.Cluster.Calico.Subnet}}/{{.Cluster.Calico.CIDR}}
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
- path: /etc/kubernetes/config/kubelet-config.yaml.tmpl
  owner: root
  permissions: 0644
  content: |
    kind: KubeletConfiguration
    apiVersion: kubelet.config.k8s.io/v1beta1
    address: ${DEFAULT_IPV4}
    port: 10250
    healthzBindAddress: ${DEFAULT_IPV4}
    healthzPort: 10248
    clusterDNS:
      - {{.Cluster.Kubernetes.DNS.IP}}
    clusterDomain: {{.Cluster.Kubernetes.Domain}}
    staticPodPath: /etc/kubernetes/manifests
    evictionSoft:
      memory.available: "500Mi"
    evictionHard:
      memory.available: "200Mi"
    evictionSoftGracePeriod:
      memory.available: "5s"
    evictionMaxPodGracePeriod: 60
    authentication:
      anonymous:
        enabled: true # Defaults to false as of 1.10
      webhook:
        enabled: false # Deafults to true as of 1.10
    authorization:
      mode: AlwaysAllow # Deafults to webhook as of 1.10
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
- path: /etc/kubernetes/config/scheduler-config.yml
  owner: root
  permissions: 0644
  content: |
    kind: KubeSchedulerConfiguration
    algorithmSource:
      provider: DefaultProvider
    apiVersion: componentconfig/v1alpha1
    clientConnection:
      kubeconfig: /etc/kubernetes/config/scheduler-kubeconfig.yml
    failureDomains: kubernetes.io/hostname,failure-domain.beta.kubernetes.io/zone,failure-domain.beta.kubernetes.io/region
    hardPodAffinitySymmetricWeight: 1
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

{{ if not .DisableEncryptionAtREST -}}
- path: /etc/kubernetes/encryption/k8s-encryption-config.yaml
  owner: root
  permissions: 600
  content: |
    kind: EncryptionConfig
    apiVersion: v1
    resources:
      - resources:
        - secrets
        providers:
        - aescbc:
            keys:
            - name: key1
              secret: {{ .APIServerEncryptionKey }}
        - identity: {}
{{ end -}}

- path: /etc/kubernetes/policies/audit-policy.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: audit.k8s.io/v1
    kind: Policy
    rules:
      # TODO: Filter safe system requests.
      # A catch-all rule to log all requests at the Metadata level.
      - level: Metadata
        # Long-running requests like watches that fall under this rule will not
        # generate an audit event in RequestReceived.
        omitStages:
          - "RequestReceived"

- path: /etc/kubernetes/manifests/k8s-api-server.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Pod
    metadata:
      name: k8s-api-server
      namespace: kube-system
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      priorityClassName: system-node-critical
      containers:
      - name: k8s-api-server
        image: {{ .RegistryDomain }}/{{ .Images.Kubernetes }}
        env:
        - name: HOST_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        command:
        - /hyperkube
        - apiserver
        {{ range .Hyperkube.Apiserver.Pod.CommandExtraArgs -}}
        - {{ . }}
        {{ end -}}
        - --allow-privileged=true
        - --insecure-bind-address=0.0.0.0
        - --anonymous-auth=false
        - --insecure-port=0
        - --kubelet-https=true
        - --kubelet-preferred-address-types=InternalIP
        - --secure-port={{.Cluster.Kubernetes.API.SecurePort}}
        - --bind-address=$(HOST_IP)
        - --etcd-prefix={{.Cluster.Etcd.Prefix}}
        - --profiling=false
        - --repair-malformed-updates=false
        - --service-account-lookup=true
        - --authorization-mode=RBAC
        - --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota,DefaultStorageClass,PersistentVolumeClaimResize,PodSecurityPolicy,Priority,DefaultTolerationSeconds,MutatingAdmissionWebhook,ValidatingAdmissionWebhook
        - --cloud-provider={{.Cluster.Kubernetes.CloudProvider}}
        - --service-cluster-ip-range={{.Cluster.Kubernetes.API.ClusterIPRange}}
        - --etcd-servers=https://127.0.0.1:2379
        - --etcd-cafile=/etc/kubernetes/ssl/etcd/server-ca.pem
        - --etcd-certfile=/etc/kubernetes/ssl/etcd/server-crt.pem
        - --etcd-keyfile=/etc/kubernetes/ssl/etcd/server-key.pem
        - --advertise-address=$(HOST_IP)
        - --runtime-config=api/all=true,scheduling.k8s.io/v1alpha1=true
        - --logtostderr=true
        - --tls-cert-file=/etc/kubernetes/ssl/apiserver-crt.pem
        - --tls-private-key-file=/etc/kubernetes/ssl/apiserver-key.pem
        - --client-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem
        - --service-account-key-file=/etc/kubernetes/ssl/service-account-key.pem
        - --audit-log-path=/var/log/apiserver/audit.log
        - --audit-log-maxage=30
        - --audit-log-maxbackup=30
        - --audit-log-maxsize=100
        - --audit-policy-file=/etc/kubernetes/policies/audit-policy.yml
        - --experimental-encryption-provider-config=/etc/kubernetes/encryption/k8s-encryption-config.yaml
        - --requestheader-client-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem
        - --requestheader-allowed-names=aggregator,{{.Cluster.Kubernetes.API.Domain}},{{.Cluster.Kubernetes.Kubelet.Domain}}
        - --requestheader-extra-headers-prefix=X-Remote-Extra-
        - --requestheader-group-headers=X-Remote-Group
        - --requestheader-username-headers=X-Remote-User
        - --proxy-client-cert-file=/etc/kubernetes/ssl/apiserver-crt.pem
        - --proxy-client-key-file=/etc/kubernetes/ssl/apiserver-key.pem
        resources:
          requests:
            cpu: 300m
            memory: 300Mi
        livenessProbe:
          tcpSocket:
            port: {{.Cluster.Kubernetes.API.SecurePort}}
          initialDelaySeconds: 15
          timeoutSeconds: 15
        ports:
        - containerPort: {{.Cluster.Kubernetes.API.SecurePort}}
          hostPort: {{.Cluster.Kubernetes.API.SecurePort}}
          name: https
        volumeMounts:
        {{ range .Hyperkube.Apiserver.Pod.HyperkubePodHostExtraMounts -}}
        - mountPath: {{ .Path }}
          name: {{ .Name }}
          readOnly: {{ .ReadOnly }}
        {{ end -}}
        - mountPath: /var/log/apiserver/
          name: apiserver-log
        - mountPath: /etc/kubernetes/encryption/
          name: k8s-encryption
          readOnly: true
        - mountPath: /etc/kubernetes/policies
          name: k8s-policies
          readOnly: true
        - mountPath: /etc/kubernetes/manifests
          name: k8s-manifests
          readOnly: true
        - mountPath: /etc/kubernetes/secrets/
          name: k8s-secrets
          readOnly: true
        - mountPath: /etc/kubernetes/ssl/
          name: ssl-certs-kubernetes
          readOnly: true
      volumes:
      {{ range .Hyperkube.Apiserver.Pod.HyperkubePodHostExtraMounts -}}
      - hostPath:
          path: {{ .Path }}
        name: {{ .Name }}
      {{ end -}}
      - hostPath:
          path: /var/log/apiserver/
        name: apiserver-log
      - hostPath:
          path: /etc/kubernetes/encryption/
        name: k8s-encryption
      - hostPath:
          path: /etc/kubernetes/policies
        name: k8s-policies
      - hostPath:
          path: /etc/kubernetes/manifests
        name: k8s-manifests
      - hostPath:
          path: /etc/kubernetes/secrets
        name: k8s-secrets
      - hostPath:
          path: /etc/kubernetes/ssl
        name: ssl-certs-kubernetes

- path: /etc/kubernetes/manifests/k8s-controller-manager.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Pod
    metadata:
      name: k8s-controller-manager
      namespace: kube-system
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      priorityClassName: system-node-critical
      containers:
      - name: k8s-controller-manager
        image: {{ .RegistryDomain }}/{{ .Images.Kubernetes }}
        command:
        - /hyperkube
        - controller-manager
        {{ range .Hyperkube.ControllerManager.Pod.CommandExtraArgs -}}
        - {{ . }}
        {{ end -}}
        - --logtostderr=true
        - --v=2
        - --cloud-provider={{.Cluster.Kubernetes.CloudProvider}}
        - --terminated-pod-gc-threshold=10
        - --use-service-account-credentials=true
        - --kubeconfig=/etc/kubernetes/config/controller-manager-kubeconfig.yml
        - --root-ca-file=/etc/kubernetes/ssl/apiserver-ca.pem
        - --service-account-private-key-file=/etc/kubernetes/ssl/service-account-key.pem
        resources:
          requests:
            cpu: 200m
            memory: 200Mi
        livenessProbe:
          httpGet:
            host: 127.0.0.1
            path: /healthz
            port: 10251
          initialDelaySeconds: 15
          timeoutSeconds: 15
        volumeMounts:
        {{ range .Hyperkube.ControllerManager.Pod.HyperkubePodHostExtraMounts -}}
        - mountPath: {{ .Path }}
          name: {{ .Name }}
          readOnly: {{ .ReadOnly }}
        {{ end -}}
        - mountPath: /etc/kubernetes/config/
          name: k8s-config
          readOnly: true
        - mountPath: /etc/kubernetes/secrets/
          name: k8s-secrets
          readOnly: true
        - mountPath: /etc/kubernetes/ssl/
          name: ssl-certs-kubernetes
          readOnly: true
      volumes:
      {{ range .Hyperkube.ControllerManager.Pod.HyperkubePodHostExtraMounts -}}
      - hostPath:
          path: {{ .Path }}
        name: {{ .Name }}
      {{ end -}}
      - hostPath:
          path: /etc/kubernetes/config
        name: k8s-config
      - hostPath:
          path: /etc/kubernetes/secrets
        name: k8s-secrets
      - hostPath:
          path: /etc/kubernetes/ssl
        name: ssl-certs-kubernetes

- path: /etc/kubernetes/manifests/k8s-scheduler.yml
  owner: root
  permissions: 0644
  content: |
    apiVersion: v1
    kind: Pod
    metadata:
      name: k8s-scheduler
      namespace: kube-system
      annotations:
        scheduler.alpha.kubernetes.io/critical-pod: ''
    spec:
      hostNetwork: true
      priorityClassName: system-node-critical
      containers:
      - name: k8s-scheduler
        image: {{ .RegistryDomain }}/{{ .Images.Kubernetes }}
        command:
        - /hyperkube
        - scheduler
        - --config=/etc/kubernetes/config/scheduler-config.yml
        - --v=2
        resources:
          requests:
            cpu: 100m
            memory: 100Mi
        livenessProbe:
          httpGet:
            host: 127.0.0.1
            path: /healthz
            port: 10251
          initialDelaySeconds: 15
          timeoutSeconds: 15
        volumeMounts:
        - mountPath: /etc/kubernetes/config/
          name: k8s-config
          readOnly: true
        - mountPath: /etc/kubernetes/ssl/
          name: ssl-certs-kubernetes
          readOnly: true
      volumes:
      - hostPath:
          path: /etc/kubernetes/config
        name: k8s-config
      - hostPath:
          path: /etc/kubernetes/ssl
        name: ssl-certs-kubernetes

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
    fs.inotify.max_user_watches = 16384
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
  - name: k8s-setup-kubelet-config.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=k8s-setup-kubelet-config Service
      After=k8s-setup-network-env.service docker.service
      Requires=k8s-setup-network-env.service docker.service

      [Service]
      EnvironmentFile=/etc/network-environment
      ExecStart=/bin/bash -c '/usr/bin/envsubst </etc/kubernetes/config/kubelet-config.yaml.tmpl >/etc/kubernetes/config/kubelet-config.yaml'

      [Install]
      WantedBy=multi-user.target
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
    enable: false
    content: |
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
      --config=/etc/kubernetes/config/kubelet-config.yaml \
      --containerized \
      --enable-server \
      --logtostderr=true \
      --cloud-provider={{.Cluster.Kubernetes.CloudProvider}} \
      --network-plugin=cni \
      --register-node=true \
      --register-with-taints=node-role.kubernetes.io/master=:NoSchedule \
      --kubeconfig=/etc/kubernetes/config/kubelet-kubeconfig.yml \
      --node-labels="node-role.kubernetes.io/master,role=master,ip=${DEFAULT_IPV4},{{.Cluster.Kubernetes.Kubelet.Labels}}" \
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
    mask: true
  - name: k8s-addons.service
    enable: true
    command: start
    content: |
      [Unit]
      Description=Kubernetes Addons
      Wants=k8s-kubelet.service
      After=k8s-kubelet.service
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

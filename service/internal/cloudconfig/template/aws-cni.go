package template

const AwsCNIManifest = `
# Vendored from https://raw.githubusercontent.com/aws/amazon-vpc-cni-k8s/master/config/v1.6/aws-k8s-cni.yaml

---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
    name: aws-cni
spec:
    allowPrivilegeEscalation: true
    privileged: true
    allowedCapabilities:
      - 'NET_ADMIN'
    fsGroup:
      rule: RunAsAny
    runAsUser:
      rule: RunAsAny
    seLinux:
      rule: RunAsAny
    supplementalGroups:
      rule: RunAsAny
    hostNetwork: true
    hostPorts:
    - min: 0
      max: 65535
    volumes:
    - secret
    - configMap
    - hostPath
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-node
rules:
- apiGroups:
  - crd.k8s.amazonaws.com
  resources:
  - eniconfigs
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - namespaces
  verbs:
  - list
  - watch
  - get
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - list
  - watch
  - get
  - update
- apiGroups:
  - extensions
  resources:
  - "*"
  verbs:
  - list
  - watch
- apiGroups: ["policy"]
  resources: ["podsecuritypolicies"]
  resourceNames: ["aws-cni"]
  verbs: ["use", "get", "create"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-node
  namespace: kube-system

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aws-node
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aws-node
subjects:
  - kind: ServiceAccount
    name: aws-node
    namespace: kube-system
---
kind: DaemonSet
apiVersion: apps/v1
metadata:
  name: aws-node
  namespace: kube-system
  labels:
    k8s-app: aws-node
spec:
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: "10%"
  selector:
    matchLabels:
      k8s-app: aws-node
  template:
    metadata:
      labels:
        k8s-app: aws-node
    spec:
      priorityClassName: system-node-critical
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
              - matchExpressions:
                  - key: "beta.kubernetes.io/os"
                    operator: In
                    values:
                      - linux
                  - key: "beta.kubernetes.io/arch"
                    operator: In
                    values:
                      - amd64
                      - arm64
                  - key: eks.amazonaws.com/compute-type
                    operator: NotIn
                    values:
                      - fargate
              - matchExpressions:
                  - key: "kubernetes.io/os"
                    operator: In
                    values:
                      - linux
                  - key: "kubernetes.io/arch"
                    operator: In
                    values:
                      - amd64
                      - arm64
                  - key: eks.amazonaws.com/compute-type
                    operator: NotIn
                    values:
                      - fargate
      serviceAccountName: aws-node
      hostNetwork: true
      tolerations:
        - operator: Exists
      initContainers:
        - image: {{.RegistryDomain}}/giantswarm/aws-cni-init:v{{.AWSCNIVersion}}
          imagePullPolicy: Always
          name: aws-vpc-cni-init
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
      containers:
        - image: {{.RegistryDomain}}/giantswarm/aws-cni:v{{.AWSCNIVersion}}
          ports:
            - containerPort: 61678
              name: metrics
          name: aws-node
          readinessProbe:
            exec:
              command: ["/app/grpc-health-probe", "-addr=:50051"]
            initialDelaySeconds: 35
          livenessProbe:
            exec:
              command: ["/app/grpc-health-probe", "-addr=:50051", "-connect-timeout=2s", "-rpc-timeout=2s" ]
            initialDelaySeconds: 35
          env:
            - name: ADDITIONAL_ENI_TAGS
              value: '{{ .AWSCNIAdditionalTags }}'
            - name: AWS_VPC_K8S_CNI_LOGLEVEL
              value: INFO
            - name: AWS_VPC_K8S_PLUGIN_LOG_LEVEL
              value: INFO
            - name: AWS_VPC_K8S_CNI_LOG_FILE
              value: stdout
            - name: AWS_VPC_K8S_PLUGIN_LOG_FILE
              value: /host/var/log/aws-routed-eni/plugin.log
            - name: AWS_VPC_ENI_MTU
              value: "9001"
            - name: AWS_VPC_K8S_CNI_CONFIGURE_RPFILTER
              value: "false"
            - name: DISABLE_INTROSPECTION
              value: "false"
            - name: DISABLE_METRICS
              value: "false"
            - name: WARM_IP_TARGET
              value: "{{ .AWSCNIWarmIPTarget }}"
            - name: MINIMUM_IP_TARGET
              value: "{{ .AWSCNIMinimumIPTarget }}"
            ## Deviation from original manifest - 1
            ## This config value is important - See here https://github.com/aws/amazon-vpc-cni-k8s/blob/master/README.md#cni-configuration-variables
            - name: AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG
              value: "true"
            ## Deviation from original manifest - 2
            ## setting custom ENI config annotation
            - name: ENI_CONFIG_LABEL_DEF
              value: "failure-domain.beta.kubernetes.io/zone"
            - name: MY_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            ## Deviation from original manifest - 3
            ## disable SNAT as we setup NATGW in the route tables
            - name: AWS_VPC_K8S_CNI_EXTERNALSNAT
              value: "{{.ExternalSNAT}}"
            {{- if eq .ExternalSNAT false }}
            ## Deviation from original manifest - 4
            ## If we left this enabled, cross subnet communication doesn't work. Only affects ExternalSNAT=false.
            - name: AWS_VPC_K8S_CNI_RANDOMIZESNAT
              value: "none"
            {{- end }}
            ## Deviation from original manifest - 5
            ## Explicit interface naming
            - name: AWS_VPC_K8S_CNI_VETHPREFIX
              value: eni
            {{- if eq .AWSCNIPrefix true }}
            ## Deviation from original manifest - 6
            ## If CNI prefix validation is enabled it will improve the speed of allocating IPs on a nitro based node (e.g. m5.2xlarge) 
            ## By setting a annotation on the Cluster CR it can be enabled or disabled.
            - name: ENABLE_PREFIX_DELEGATION
              value: "true"
            - name: "WARM_PREFIX_TARGET"
              value: "1"
            {{- end }}
          resources:
            requests:
              cpu: 30m
          securityContext:
            capabilities:
              add:
                - NET_ADMIN
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
            - mountPath: /host/var/log/aws-routed-eni
              name: log-dir
            - mountPath: /var/run/aws-node
              name: run-dir
            - mountPath: /var/run/dockershim.sock
              name: dockershim
            - mountPath: /run/xtables.lock
              name: xtables-lock
      volumes:
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        - name: dockershim
          hostPath:
            path: /var/run/dockershim.sock
        - hostPath:
            path: /run/xtables.lock
          name: xtables-lock
        - hostPath:
            path: /var/log/aws-routed-eni
            type: DirectoryOrCreate
          name: log-dir
        - hostPath:
            path: /var/run/aws-node
            type: DirectoryOrCreate
          name: run-dir
---
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: eniconfigs.crd.k8s.amazonaws.com
spec:
  scope: Cluster
  group: crd.k8s.amazonaws.com
  versions:
    - name: v1alpha1
      served: true
      storage: true
  names:
    plural: eniconfigs
    singular: eniconfig
    kind: ENIConfig
---
## AWS CNI restarter, to be removed when AWS CNI is able to detect new Additional CIDRs https://github.com/giantswarm/giantswarm/issues/11077
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-cni-restarter
  namespace: kube-system
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: aws-cni-restarter
  namespace: kube-system
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "create", "patch"]
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["list", "delete"]
- apiGroups: ["policy"]
  resources: ["podsecuritypolicies"]
  resourceNames: ["aws-cni-restarter"]
  verbs: ["use", "get", "create"]
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: aws-cni-restarter-binding
  namespace: kube-system
subjects:
- kind: ServiceAccount
  name: aws-cni-restarter
  namespace: kube-system
roleRef:
  kind: Role
  name: aws-cni-restarter
  apiGroup: ""
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
    name: aws-cni-restarter
spec:
  fsGroup:
    rule: RunAsAny
  hostNetwork: true
  privileged: false
  runAsUser:
    rule: MustRunAsNonRoot
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
    - secret
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  labels:
    app: aws-cni-restarter
  name: aws-cni-restarter
  namespace: kube-system
spec:
  egress:
  - {}
  podSelector:
    matchLabels:
      app: aws-cni-restarter
  policyTypes:
  - Egress
---
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: aws-cni-restarter
  namespace: kube-system
spec:
  suspend: true
  schedule: "*/5 * * * *"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 5
  failedJobsHistoryLimit: 10
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app: aws-cni-restarter
        spec:
          serviceAccountName: aws-cni-restarter
          hostNetwork: true
          containers:
            - name: aws-cni-restarter
              image: {{.RegistryDomain}}/giantswarm/aws-cni-restarter:1.0.2
          restartPolicy: OnFailure
          affinity:
            nodeAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                nodeSelectorTerms:
                  - matchExpressions:
                    - key: role
                      operator: In
                      values:
                      - master
          tolerations:
          - effect: NoSchedule
            key: node-role.kubernetes.io/master
`

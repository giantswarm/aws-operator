package template

const AwsCloudControllerManagerManifest = `

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aws-cloud-controller-manager
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-cloud-controller-manager
rules:
- apiGroups:
  - extensions
  resources:
  - podsecuritypolicies
  resourceNames:
  - aws-cloud-controller-manager
  verbs:
  - use
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
  - update
- apiGroups:
  - ""
  resources:
  - nodes
  verbs:
  - '*'
- apiGroups:
  - ""
  resources:
  - nodes/status
  verbs:
  - patch
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - services/status
  verbs:
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - persistentvolumes
  verbs:
  - get
  - list
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - endpoints
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - coordination.k8s.io
  resources:
  - leases
  verbs:
  - create
  - get
  - list
  - watch
  - update
- apiGroups:
  - ""
  resources:
  - serviceaccounts/token
  verbs:
  - create
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - list
  - watch
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: aws-cloud-controller-manager
subjects:
- kind: ServiceAccount
  name: aws-cloud-controller-manager
  namespace: kube-system
roleRef:
  kind: ClusterRole
  name: aws-cloud-controller-manager
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: aws-cloud-controller-manager
  namespace: kube-system
spec:
  privileged: false
  allowPrivilegeEscalation: false
  allowedCapabilities: []
  volumes:
    - 'hostPath'
    - 'projected'
  hostNetwork: true
  hostIPC: false
  hostPID: false
  hostPorts: []
  runAsUser:
    rule: 'RunAsAny'
  seLinux:
    rule: 'RunAsAny'
  supplementalGroups:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
  fsGroup:
    rule: 'MustRunAs'
    ranges:
      - min: 1
        max: 65535
---
kind: NetworkPolicy
apiVersion: networking.k8s.io/v1
metadata:
  name: aws-cloud-controller-manager
  namespace: kube-system
spec:
  podSelector:
    matchLabels:
      k8s-app: aws-cloud-controller-manager
  ingress:
  - {}
  egress:
  - {}
  policyTypes:
  - Egress
  - Ingress
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: aws-cloud-controller-manager
  namespace: kube-system
  labels:
    k8s-app: aws-cloud-controller-manager
spec:
  updateStrategy:
    type: RollingUpdate
  selector:
    matchLabels:
	  k8s-app: aws-cloud-controller-manager
  template:
    metadata:
      labels:
        k8s-app: aws-cloud-controller-manager
    spec:
      priorityClassName: system-node-critical
      securityContext:
        runAsUser: 0
        runAsGroup: 0
      serviceAccountName: aws-cloud-controller-manager
      hostNetwork: true
      tolerations:
      - operator: "Exists"
      nodeSelector:
        node-role.kubernetes.io/master: ""
      containers:
      - name: aws-cloud-controller-manager
        image: "{{.RegistryDomain}}/giantswarm/aws-cloud-controller-manager:v1.23.2"
        resources:
          limits:
            cpu: 200m
            memory: 300Mi
          requests:
            cpu: 200m
            memory: 300Mi
        args:
        - --cloud-provider=aws
        - --port=10267
        - --configure-cloud-routes=false
        - --v=2
        securityContext:
          allowPrivilegeEscalation: false
          privileged: false
        readinessProbe:
          httpGet:
            host: 127.0.0.1
            path: /healthz
            port: 10267
          initialDelaySeconds: 20
          periodSeconds: 10
          timeoutSeconds: 5
        livenessProbe:
          httpGet:
            host: 127.0.0.1
            path: /healthz
            port: 10267
          initialDelaySeconds: 20
          periodSeconds: 10
          timeoutSeconds: 5
`

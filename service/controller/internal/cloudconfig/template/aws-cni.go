package template

const AwsCNIManifest = `
# Vendored from https://raw.githubusercontent.com/aws/amazon-vpc-cni-k8s/master/config/v1.6/aws-k8s-cni.yaml

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aws-node
rules:
  - apiGroups:
      - crd.k8s.amazonaws.com
    resources:
      - "*"
      - namespaces
    verbs:
      - "*"
  - apiGroups: [""]
    resources:
      - pods
      - nodes
      - namespaces
    verbs: ["list", "watch", "get"]
  - apiGroups: ["extensions"]
    resources:
      - daemonsets
    verbs: ["list", "watch"]
  ## Deviation from original manifest - 1
  ## add RBAC rules to use privileged PSP
  - apiGroups:
      - extensions
    resources:
      - podsecuritypolicies
    resourceNames:
      - privileged
    verbs:
      - use

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
                  - key: eks.amazonaws.com/compute-type
                    operator: NotIn
                    values:
                      - fargate
      serviceAccountName: aws-node
      hostNetwork: true
      tolerations:
        - operator: Exists
      containers:
        - image: {{.RegistryDomain}}/giantswarm/amazon-k8s-cni:v1.6.0
          imagePullPolicy: Always
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
              command: ["/app/grpc-health-probe", "-addr=:50051"]
            initialDelaySeconds: 35
          env:
            - name: AWS_VPC_K8S_CNI_LOGLEVEL
              value: DEBUG
            - name: AWS_VPC_ENI_MTU
              value: "9001"
            ## Deviation from original manifest - 2
            ## This config value is important - See here https://github.com/aws/amazon-vpc-cni-k8s/blob/master/README.md#cni-configuration-variables
            - name: AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG
              value: "true"
            ## Deviation from original manifest - 3
            ## setting custom ENI config annotation
            - name: ENI_CONFIG_LABEL_DEF
              value: "failure-domain.beta.kubernetes.io/zone"
            - name: MY_NODE_NAME
              valueFrom:
                fieldRef:
                  fieldPath: spec.nodeName
            ## Deviation from original manifest - 4
            ## prewarm IP to limit AWS api calls
            - name: WARM_IP_TARGET
              value: "20"
            - name: MINIMUM_IP_TARGET
              value: "5"
            ## Deviation from original manifest - 5
            ## disable SNAT as we setup NATGW in the route tables
            - name: AWS_VPC_K8S_CNI_EXTERNALSNAT
              value: "{{.ExternalSNAT}}"
            {{- if .ExternalSNAT }}
            ## If we left this enabled, cross subnet communication doesn't work. Only affects EcternalSnat to false
            - name: AWS_VPC_K8S_CNI_RANDOMIZESNAT
              value: "none"
            {{- end }}
            ## Deviation from original manifest - 6
            ## Explicit interface naming
            - name: AWS_VPC_K8S_CNI_VETHPREFIX
              value: eni
          resources:
            requests:
              cpu: 10m
          securityContext:
            privileged: true
          volumeMounts:
            - mountPath: /host/opt/cni/bin
              name: cni-bin-dir
            - mountPath: /host/etc/cni/net.d
              name: cni-net-dir
            - mountPath: /host/var/log
              name: log-dir
            - mountPath: /var/run/docker.sock
              name: dockersock
            - mountPath: /var/run/dockershim.sock
              name: dockershim
      volumes:
        - name: cni-bin-dir
          hostPath:
            path: /opt/cni/bin
        - name: cni-net-dir
          hostPath:
            path: /etc/cni/net.d
        - name: log-dir
          hostPath:
            path: /var/log
        - name: dockersock
          hostPath:
            path: /var/run/docker.sock
        - name: dockershim
          hostPath:
            path: /var/run/dockershim.sock

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



`

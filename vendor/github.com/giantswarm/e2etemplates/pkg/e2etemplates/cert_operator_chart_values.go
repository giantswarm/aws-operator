package e2etemplates

// CertOperatorChartValues values required by cert-operator-chart, the environment
// variables will be expanded before writing the contents to a file.
const CertOperatorChartValues = `commonDomain: k8s.${COMMON_DOMAIN}
clusterName: ${CLUSTER_NAME}
Installation:
  V1:
    Auth:
      Vault:
        Address: http://vault.default.svc.cluster.local:8200
        CA:
          TTL: 1440h
    Guest:
      Kubernetes:
        API:
          EndpointBase: k8s.${COMMON_DOMAIN}
    Secret:
      CertOperator:
        SecretYaml: |
          service:
            vault:
              config:
                token: ${VAULT_TOKEN}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"$REGISTRY_PULL_SECRET\"}}}"
labelSelector: 'giantswarm.io/cluster=${CLUSTER_NAME}'
`

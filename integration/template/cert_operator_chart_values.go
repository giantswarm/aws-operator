package template

// CertOperatorChartValues values required by cert-operator-chart, the environment
// variables will be expanded before writing the contents to a file.
var CertOperatorChartValues = `commonDomain: ${COMMON_DOMAIN_GUEST}
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
          EndpointBase: ${COMMON_DOMAIN_GUEST}
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
`

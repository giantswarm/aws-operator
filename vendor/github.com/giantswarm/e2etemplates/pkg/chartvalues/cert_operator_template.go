package chartvalues

const certOperatorTemplate = `clusterRoleBindingName: {{ .ClusterRole.BindingName }}
clusterRoleBindingNamePSP: {{ .ClusterRolePSP.BindingName }}
clusterRoleName: {{ .ClusterRole.Name }}
clusterRoleNamePSP: {{ .ClusterRolePSP.Name }}
Installation:
  V1:
    Auth:
      Vault:
        Address: http://vault.default.svc.cluster.local:8200
        CA:
          TTL: 1440h
    GiantSwarm:
      CertOperator:
        CRD:
          LabelSelector: 'giantswarm.io/cluster={{ .ClusterName }}'
    Guest:
      Kubernetes:
        API:
          EndpointBase: {{ .CommonDomain }}
    Secret:
      CertOperator:
        SecretYaml: |
          service:
            vault:
              config:
                token: {{ .Vault.Token }}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
pspName: {{ .PSP.Name }}
`

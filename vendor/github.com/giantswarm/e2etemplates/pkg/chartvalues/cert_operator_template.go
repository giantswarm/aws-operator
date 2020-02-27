package chartvalues

const certOperatorTemplate = `
clusterRoleBindingName: {{ .ClusterRole.BindingName }}
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
          LabelSelector: {{ .CRD.LabelSelector }}
    Guest:
      Kubernetes:
        API:
          EndpointBase: k8s.{{ .CommonDomain }}
    Secret:
      CertOperator:
        Service:
          Vault:
            Config:
              Token: {{ .Vault.Token }}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
namespace: {{ .Namespace }}
pspName: {{ .PSP.Name }}
`

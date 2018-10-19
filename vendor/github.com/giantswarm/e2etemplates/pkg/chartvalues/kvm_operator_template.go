package chartvalues

const kvmOperatorTemplate = `
clusterRoleBindingName: {{ .ClusterRole.BindingName }}
clusterRoleBindingNamePSP: {{ .ClusterRolePSP.BindingName }}
clusterRoleName: {{ .ClusterRole.Name }}
clusterRoleNamePSP: {{ .ClusterRolePSP.Name }}
Installation:
  V1:
    GiantSwarm:
      KVMOperator:
        CRD:
          LabelSelector: 'giantswarm.io/cluster={{ .ClusterName }}'
    Guest:
      SSH:
        SSOPublicKey: 'test'
      Kubernetes:
        API:
          Auth:
            Provider:
              OIDC:
                ClientID: ""
                IssueURL: ""
                UsernameClaim: ""
                GroupsClaim: ""
      Update:
        Enabled: true
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
namespace: {{ .Namespace }}
pspName: {{ .PSP.Name }}
`

package chartvalues

const flannelOperatorTemplate = `clusterRoleBindingName: {{ .ClusterRole.BindingName }}
clusterRoleBindingNamePSP: {{ .ClusterRolePSP.BindingName }}
clusterRoleName: {{ .ClusterRole.Name }}
clusterRoleNamePSP: {{ .ClusterRolePSP.Name }}
Installation:
  V1:
    GiantSwarm:
      FlannelOperator:
        CRD:
          LabelSelector: 'giantswarm.io/cluster={{ .ClusterName }}'
    Secret:
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
pspName: {{ .PSP.Name }}
`

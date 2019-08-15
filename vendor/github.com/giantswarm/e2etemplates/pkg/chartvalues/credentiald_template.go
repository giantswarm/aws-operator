package chartvalues

const credentialdTemplate = `
deployment:
  replicas: {{ .Deployment.Replicas }}
Installation:
  V1:
    Secret:
      Credentiald:
      {{- if .AWS.CredentialDefault.AWSOperatorARN }}
        AWS:
          CredentialDefault:
            AdminARN: "{{ .AWS.CredentialDefault.AdminARN }}"
            AWSOperatorARN: "{{ .AWS.CredentialDefault.AWSOperatorARN }}"
      {{- end }}
      {{- if .Azure.CredentialDefault.ClientID }}
        Azure:
          CredentialDefault:
            ClientID: "{{ .Azure.CredentialDefault.ClientID }}"
            ClientSecret: "{{ .Azure.CredentialDefault.ClientSecret }}"
            SubscriptionID: "{{ .Azure.CredentialDefault.SubscriptionID }}"
            TenantID: "{{ .Azure.CredentialDefault.TenantID }}"
      {{- end }}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
`

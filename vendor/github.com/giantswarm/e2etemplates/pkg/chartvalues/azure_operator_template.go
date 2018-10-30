package chartvalues

const azureOperatorTemplate = `
Installation:
  V1:
    Auth:
      Vault:
        Address: http://vault.default.svc.cluster.local:8200
    Guest:
      IPAM:
        NetworkCIDR: "10.12.0.0/16"
        CIDRMask: 24
        PrivateSubnetMask: 25
        PublicSubnetMask: 25
      Kubernetes:
        API:
          Auth:
            Provider:
              OIDC:
                ClientID: ""
                IssueURL: ""
                UsernameClaim: ""
                GroupsClaim: ""
      SSH:
        SSOPublicKey: 'test'
      Update:
        Enabled: true
    Name: ci-azure-operator
    Provider:
      Azure:
        # TODO rename to EnvironmentName. See https://github.com/giantswarm/giantswarm/issues/4124.
        Cloud: AZUREPUBLICCLOUD
        HostCluster:
          CIDR: "10.0.0.0/16"
          ResourceGroup: "godsmack"
          VirtualNetwork: "godsmack"
          VirtualNetworkGateway: "godsmack-vpn-gateway"
        MSI:
          Enabled: true
        Location: {{ .Provider.Azure.Location }}
    Registry:
      Domain: quay.io
    Secret:
      AzureOperator:
        CredentialDefault:
          clientid: {{ .Secret.AzureOperator.CredentialDefault.ClientID }}
          clientsecret: {{ .Secret.AzureOperator.CredentialDefault.ClientSecret }}
          subscriptionid: {{ .Secret.AzureOperator.CredentialDefault.SubscriptionID }}
          tenantid: {{ .Secret.AzureOperator.CredentialDefault.TenantID }}
        SecretYaml: |
          service:
            azure:
              clientid: {{ .Secret.AzureOperator.SecretYaml.Service.Azure.ClientID }}
              clientsecret: {{ .Secret.AzureOperator.SecretYaml.Service.Azure.ClientSecret }}
              subscriptionid: {{ .Secret.AzureOperator.SecretYaml.Service.Azure.SubscriptionID }}
              tenantid: {{ .Secret.AzureOperator.SecretYaml.Service.Azure.TenantID }}
              template:
                uri:
                  version: {{ .Secret.AzureOperator.SecretYaml.Service.Azure.Template.URI.Version }}
      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .Secret.Registry.PullSecret.DockerConfigJSON }}\"}}}"
    Security:
      RestrictAccess:
        Enabled: false
`

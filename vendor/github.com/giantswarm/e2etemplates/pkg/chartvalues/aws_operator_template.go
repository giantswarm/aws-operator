package chartvalues

const awsOperatorTemplate = `Installation:
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
    Name: ci-aws-operator
    Provider:
      AWS:
        AvailabilityZones:
          - eu-central-1a
          - eu-central-1b
          - eu-central-1c
        Region: '{{ .Provider.AWS.Region }}'
        DeleteLoggingBucket: true
        IncludeTags: true
        Route53:
          Enabled: true
        RouteTableNames: 'gauss_private_0,gauss_private_1'
        Encrypter: '{{ .Provider.AWS.Encrypter }}'
        TrustedAdvisor:
          Enabled: false
    Registry:
      Domain: quay.io
    Secret:
      AWSOperator:
        CredentialDefault:
          AdminARN: '{{ .Secret.AWSOperator.CredentialDefault.AdminARN }}'
          AWSOperatorARN: '{{ .Secret.AWSOperator.CredentialDefault.AWSOperatorARN }}'
        IDRSAPub: {{ .Secret.AWSOperator.IDRSAPub }}
        SecretYaml: |
          service:
            aws:
              accesskey:
                id: '{{ .Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.ID }}'
                secret: '{{ .Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.Secret }}'
                token: '{{ .Secret.AWSOperator.SecretYaml.Service.AWS.AccessKey.Token }}'
              hostaccesskey:
                id: '{{ .Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.ID }}'
                secret: '{{ .Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.Secret }}'
                token: '{{ .Secret.AWSOperator.SecretYaml.Service.AWS.HostAccessKey.Token }}'

      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"{{ .RegistryPullSecret }}\"}}}"
    Security:
      RestrictAccess:
        Enabled: false
`

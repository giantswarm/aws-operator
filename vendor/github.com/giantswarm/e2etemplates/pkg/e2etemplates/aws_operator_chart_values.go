package e2etemplates

// AWSOperatorChartValues values required by aws-operator-chart, the environment
// variables will be expanded before writing the contents to a file.
const AWSOperatorChartValues = `Installation:
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
        Enabled: ${GUEST_UPDATE_ENABLED}
    Name: ci-aws-operator
    Provider:
      AWS:
        Region: ${AWS_REGION}
        DeleteLoggingBucket: true
        IncludeTags: true
        Route53:
          Enabled: true
        Encrypter: 'kms'
    Registry:
      Domain: quay.io
    Secret:
      AWSOperator:
        IDRSAPub: ${IDRSA_PUB}
        SecretYaml: |
          service:
            aws:
              accesskey:
                id: ${GUEST_AWS_ACCESS_KEY_ID}
                secret: ${GUEST_AWS_SECRET_ACCESS_KEY}
                token: ${GUEST_AWS_SESSION_TOKEN}
              hostaccesskey:
                id: ${HOST_AWS_ACCESS_KEY_ID}
                secret: ${HOST_AWS_SECRET_ACCESS_KEY}
                token: ${HOST_AWS_SESSION_TOKEN}

      Registry:
        PullSecret:
          DockerConfigJSON: "{\"auths\":{\"quay.io\":{\"auth\":\"${REGISTRY_PULL_SECRET}\"}}}"
    Security:
      RestrictAccess:
        Enabled: false
`

package template

const TemplateMainSubnets = `
{{ define "subnets" }}
  {{ range .Subnets }}
  {{ .Name }}:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      CidrBlock: {{ .CIDR }}
      MapPublicIpOnLaunch: false
      Tags:
      - Key: Name
        Value: {{ .Name }}
      - Key: "kubernetes.io/role/elb"
        Value: "1"
      VpcId: {{ .TenantCluster.VPC.ID }}
  {{ end }}
{{ end }}
`

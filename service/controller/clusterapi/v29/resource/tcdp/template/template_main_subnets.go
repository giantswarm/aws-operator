package template

const TemplateMainSubnets = `
{{ define "subnets" }}
  {{ range .Subnets.List }}
  Subnet-{{ .NameSuffix }}:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      CidrBlock: {{ .CIDR }}
      MapPublicIpOnLaunch: false
      Tags:
      - Key: Name
        Value: Subnet-{{ .NameSuffix }}
      - Key: "kubernetes.io/role/elb"
        Value: "1"
      VpcId: {{ .TCCP.VPC.ID }}

  RouteTableAssociation-{{ .RouteTableAssociation.NameSuffix }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: {{ .TCCP.Subnet.RouteTable.ID }}
      SubnetId: {{ .TCCP.Subnet.ID }}
  {{ end }}
{{ end }}
`

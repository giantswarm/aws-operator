package template

const TemplateMainSubnets = `
{{- define "subnets" -}}
  {{- range .Subnets.List }}
  {{ .Name }}:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      CidrBlock: {{ .CIDR }}
      MapPublicIpOnLaunch: false
      Tags:
      - Key: Name
        Value: {{ .Name }}
      VpcId: {{ .TCCP.VPC.ID }}
    DependsOn: VpcCidrBlock
  {{ .RouteTableAssociation.Name }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTable.Name }}
      SubnetId: !Ref {{ .Name }}
  {{- end }}
{{- end -}}
`

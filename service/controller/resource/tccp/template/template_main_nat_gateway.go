package template

const TemplateMainNatGateway = `
{{- define "nat_gateway" -}}
  {{- $v := .NATGateway -}}
  {{- range $v.Gateways }}
  {{ .NATGWName }}:
    Type: AWS::EC2::NatGateway
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      AllocationId:
        Fn::GetAtt:
        - {{ .NATEIPName }}
        - AllocationId
      SubnetId: !Ref {{ .PublicSubnetName }}
      Tags:
        - Key: giantswarm.io/availability-zone
          Value: {{ .AvailabilityZone }}
  {{ .NATEIPName }}:
    Type: AWS::EC2::EIP
    Properties:
      Domain: vpc
  {{- end -}}
  {{- range $v.NATRoutes }}
  {{ .NATRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .PrivateRouteTableName }}
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId:
        Ref: {{ .NATGWName }}
  {{- end -}}
{{- end -}}
`

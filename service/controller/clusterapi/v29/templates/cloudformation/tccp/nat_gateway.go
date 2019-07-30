package tccp

const NatGateway = `
{{- define "nat_gateway" -}}
  {{- $v := .Guest.NATGateway -}}
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
        - Key: Name
          Value: {{ .ClusterID }}
        - Key: giantswarm.io/tccp
          Value: true
  {{ .NATEIPName }}:
    Type: AWS::EC2::EIP
    Properties:
      Domain: vpc
  Private{{ .NATRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .PrivateRouteTableName }}
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId:
        Ref: "{{ .NATGWName }}"
  Public{{ .NATRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .PublicRouteTableName }}
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId:
        Ref: "{{ .NATGWName }}"
  {{- end -}}
{{- end -}}
`

package template

const TemplateMainRouteTable = `
{{- define "routetables" -}}
  {{- range .RouteTable.List }}
  {{ .Name }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: {{ .VPC.ID }}
      Tags:
      - Key: Name
        Value: {{ .TagName }}
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: private
      - Key: giantswarm.io/tccp
        Value: true
  {{ .NATRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .PrivateRouteTableName }}
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId: {{ .NATGW.ID }}
{{- end -}}
`

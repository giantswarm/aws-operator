package template

const TemplateMainRouteTables = `
{{- define "routetables" -}}
  {{- range .RouteTables.List -}}
  {{ .Name }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: {{ .TCCP.VPC.ID }}
      Tags:
      - Key: Name
        Value: {{ .Name }}
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: private
      - Key: giantswarm.io/tccp
        Value: true
  {{ .Route.Name }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .Name }}
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId: {{ .TCCP.NATGateway.ID }}
	{{- end }}
{{- end -}}
`

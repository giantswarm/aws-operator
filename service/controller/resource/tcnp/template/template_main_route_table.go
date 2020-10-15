package template

const TemplateMainRouteTables = `
{{- define "route_tables" -}}
  {{- range .RouteTables.List }}
  {{ .Name }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: {{ .TCCP.VPC.ID }}
      Tags:
      - Key: Name
        Value: {{ .ClusterID }}-private-{{ .NodePoolID }}
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: private
  {{ .Route.Name }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .Name }}
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId: {{ .TCCP.NATGateway.ID }}
  {{- end }}
{{- end -}}
`

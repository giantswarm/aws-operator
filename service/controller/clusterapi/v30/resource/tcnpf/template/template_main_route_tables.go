package template

const TemplateMainRouteTables = `
{{- define "route_tables" -}}
  {{- range .RouteTables.PeeringConnections }}
  {{ .Name }}:
    Type: AWS::EC2::Route
    Properties:
      DestinationCidrBlock: {{ .Subnet.CIDR }}
      RouteTableId: {{ .RouteTable.ID }}
      VpcPeeringConnectionId: {{ .ID }}
  {{- end }}
{{- end -}}
`

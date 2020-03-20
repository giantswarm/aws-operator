package template

const TemplateMainRouteTables = `
{{- define "route_tables" -}}
{{- $v := .RouteTables -}} 
  {{- range $v.PrivateRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: {{ $v.VPCID }}
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}-private
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: private
  {{ .VPCPeeringRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .ResourceName }}
      DestinationCidrBlock: {{ $v.HostClusterCIDR }}
      VpcPeeringConnectionId: {{ $v.PeeringConnectionID }}
  {{- end }}
{{- end -}}
`

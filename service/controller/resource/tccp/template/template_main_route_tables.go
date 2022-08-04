package template

const TemplateMainRouteTables = `
{{- define "route_tables" -}}
{{- $v := .RouteTables -}}
  {{- if .EnableAWSCNI }}
  {{- range $v.AWSCNIRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}-aws-cni-{{ .AvailabilityZoneRegion }}
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: aws-cni
  {{- end }}
  {{- end }}
  {{- range $v.PublicRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}-public-{{ .AvailabilityZoneRegion }}
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: public
  {{- end }}
  {{- range $v.PrivateRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}-private-{{ .AvailabilityZoneRegion }}
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: private
  {{ .VPCPeeringRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .ResourceName }}
      DestinationCidrBlock: {{ $v.HostClusterCIDR }}
      VpcPeeringConnectionId:
        Ref: VPCPeeringConnection
  {{- end }}
{{- end -}}
`

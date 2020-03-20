package template

const TemplateMainRouteTables = `
{{- define "route_tables" -}}
{{- $v := .RouteTables -}}
  {{- range $v.AWSCNIRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}-aws-cni
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: aws-cni
  {{- end }}
  {{- range $v.PublicRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}-public
      - Key: giantswarm.io/availability-zone
        Value: {{ .AvailabilityZone }}
      - Key: giantswarm.io/route-table-type
        Value: public
  {{- end }}
{{- end -}}
`

package tccp

const RouteTables = `
{{ define "route_tables" }}
{{- $v := .Guest.RouteTables }}
  {{ $v.PublicRouteTableName.ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.PublicRouteTableName.TagName }}
      - Key: giantswarm.io/tccp
        Value: true

  {{- range $v.PrivateRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ .TagName }}
      - Key: giantswarm.io/tccp
        Value: true

  {{ .VPCPeeringRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .ResourceName }}
      DestinationCidrBlock: {{ $v.HostClusterCIDR }}
      VpcPeeringConnectionId:
        Ref: "VPCPeeringConnection"
  {{ end }}
{{ end }}
`

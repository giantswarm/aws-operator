package guest

const RouteTables = `{{ define "route_tables" }}
{{- $v := .Guest.RouteTables }}
  {{ $v.PublicRouteTableName.ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ $v.PublicRouteTableName.InternalName }}

  {{- range $v.PrivateRouteTableNames }}
  {{ .ResourceName }}:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ .InternalName }}

  {{ .VPCPeeringRouteName }}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref {{ .ResourceName }}
      DestinationCidrBlock: {{ $v.HostClusterCIDR }}
      VpcPeeringConnectionId:
        Ref: "VPCPeeringConnection"
  {{ end }}
{{ end }}`

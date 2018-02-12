package guest

const RouteTables = `
{{define "route_tables"}}
  PublicRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ .PublicRouteTableName }}

  PrivateRouteTable:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ .PrivateRouteTableName }}

  VPCPeeringRoute:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref PrivateRouteTable
      DestinationCidrBlock: {{ .HostClusterCIDR }}
      VpcPeeringConnectionId:
        Ref: "VPCPeeringConnection"
{{end}}
`

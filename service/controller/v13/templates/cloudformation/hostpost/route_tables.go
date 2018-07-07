package hostpost

const RouteTables = `{{define "route_tables"}}
  {{ range $i, $v := .PrivateRouteTables }}
  PrivateRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$v.RouteTableID}}
      DestinationCidrBlock: {{$v.CidrBlock}}
      VpcPeeringConnectionId: {{$v.PeerConnectionID}}
  {{end}}

  {{ range $i, $v := .PublicRouteTables }}
  PublicRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$v.RouteTableID}}
      DestinationCidrBlock: {{$v.CidrBlock}}
      VpcPeeringConnectionId: {{$v.PeerConnectionID}}
  {{ end }}

{{end}}`

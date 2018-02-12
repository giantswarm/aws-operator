package hostpost

const RouteTables = `{{define "route_tables"}}
  {{ range $i, $v := .RouteTables }}
  PrivateRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$v.RouteTableID}}
      DestinationCidrBlock: {{$v.CidrBlock}}
      VpcPeeringConnectionId: {{$v.PeerConnectionID}}
  {{end}}
{{end}}`

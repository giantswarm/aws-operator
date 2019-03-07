package cpf

const RouteTables = `
{{ define "route_tables" }}
  {{ range $i, $r := .RouteTables.PrivateRoutes }}
  PrivateRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$r.RouteTableID}}
      DestinationCidrBlock: {{$r.CidrBlock}}
      VpcPeeringConnectionId: {{$r.PeerConnectionID}}
  {{end}}

  {{ range $i, $r := .RouteTables.PublicRoutes }}
  PublicRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$r.RouteTableID}}
      DestinationCidrBlock: {{$r.CidrBlock}}
      VpcPeeringConnectionId: {{$r.PeerConnectionID}}
  {{ end }}
{{ end }}
`

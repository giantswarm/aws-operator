package hostpost

const RouteTables = `{{ define "route_tables" }}
  {{- $v := .HostPost.RouteTables }}
  {{ range $i, $t := $v.PrivateRouteTables }}
  PrivateRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$t.RouteTableID}}
      DestinationCidrBlock: {{$t.CidrBlock}}
      VpcPeeringConnectionId: {{$t.PeerConnectionID}}
  {{end}}

  {{ range $i, $t := $v.PublicRouteTables }}
  PublicRoute{{$i}}:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: {{$t.RouteTableID}}
      DestinationCidrBlock: {{$t.CidrBlock}}
      VpcPeeringConnectionId: {{$t.PeerConnectionID}}
  {{ end }}

{{ end }}`

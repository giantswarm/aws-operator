package template

const TemplateMainRouteTableAssociation = `
{{ define "route_table_association" }}
  {{ .RouteTableAssociation.Name }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTable.Name }} # tccp private subnet
      SubnetId: !Ref {{ .RouteTableAssociation.Subnet.Name }}
{{ end }}
`

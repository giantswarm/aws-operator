package tccp

const Subnets = `
{{ define "subnets" }}
{{- $v := .Guest.Subnets }}
  {{- range $v.PublicSubnets }}
  {{ .Name }}:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      CidrBlock: {{ .CIDR }}
      MapPublicIpOnLaunch: {{ .MapPublicIPOnLaunch }}
      Tags:
      - Key: Name
        Value: {{ .Name }}
      - Key: "kubernetes.io/role/elb"
        Value: "1"
      VpcId: !Ref VPC

  {{ .RouteTableAssociation.Name }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTableName }}
      SubnetId: !Ref {{ .RouteTableAssociation.SubnetName }}

  {{ end }}

  {{- range $v.PrivateSubnets }}
  {{ .Name }}:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      CidrBlock: {{ .CIDR }}
      MapPublicIpOnLaunch: {{ .MapPublicIPOnLaunch }}
      Tags:
      - Key: Name
        Value: {{ .Name }}
      - Key: "kubernetes.io/role/internal-elb"
        Value: "1"
      VpcId: !Ref VPC

  {{ .RouteTableAssociation.Name  }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTableName }}
      SubnetId: !Ref {{ .RouteTableAssociation.SubnetName }}
  {{ end }}
{{ end }}
`

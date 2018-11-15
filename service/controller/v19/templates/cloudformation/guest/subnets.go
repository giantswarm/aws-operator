package guest

const Subnets = `{{ define "subnets" }}
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
      VpcId: !Ref VPC

  PublicRouteTableAssociation{{ .Index }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref PublicRouteTable
      SubnetId: !Ref {{ .Name }}

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
      VpcId: !Ref VPC

  PrivateRouteTableAssociation{{ .Index }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref PrivateRouteTable{{ .Index }}
      SubnetId: !Ref {{ .Name }}
  {{ end }}
{{ end }}`

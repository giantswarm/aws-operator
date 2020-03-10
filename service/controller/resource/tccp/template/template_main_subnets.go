package template

const TemplateMainSubnets = `
{{- define "subnets" -}}
{{- $v := .Subnets }}
  {{- range $v.AWSCNISubnets }}
  {{ .Name }}:
    Type: AWS::EC2::Subnet
    DependsOn:
    - VPCCIDRBlockAWSCNI
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      CidrBlock: {{ .CIDR }}
      Tags:
      - Key: Name
        Value: {{ .Name }}
      - Key: giantswarm.io/subnet-type
        Value: aws-cni
      VpcId: !Ref VPC
  {{ .RouteTableAssociation.Name }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTableName }}
      SubnetId: !Ref {{ .RouteTableAssociation.SubnetName }}
  {{- end }}
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
      - Key: giantswarm.io/subnet-type
        Value: public
      - Key: kubernetes.io/role/elb
        Value: 1
      VpcId: !Ref VPC
  {{ .RouteTableAssociation.Name }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTableName }}
      SubnetId: !Ref {{ .RouteTableAssociation.SubnetName }}
  {{- end }}
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
      - Key: giantswarm.io/subnet-type
        Value: private
      - Key: kubernetes.io/role/internal-elb
        Value: 1
      VpcId: !Ref VPC
  {{ .RouteTableAssociation.Name  }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTableName }}
      SubnetId: !Ref {{ .RouteTableAssociation.SubnetName }}
  {{- end }}
{{- end -}}
`

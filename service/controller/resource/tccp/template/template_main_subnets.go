package template

const TemplateMainSubnets = `
{{- define "subnets" -}}
{{- $v := .Subnets }}
  {{- if .EnableAWSCNI }}
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
      - Key: sigs.k8s.io/cluster-api-provider-aws/role
        Value: public
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
      - Key: sigs.k8s.io/cluster-api-provider-aws/role
        Value: private
      VpcId: !Ref VPC
  {{ .RouteTableAssociation.Name  }}:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref {{ .RouteTableAssociation.RouteTableName }}
      SubnetId: !Ref {{ .RouteTableAssociation.SubnetName }}
  {{- end }}
{{- end -}}
`

package guest

const Subnets = `{{ define "subnets" }}
{{- $v := .Guest.Subnets }}
  PublicSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ $v.PublicSubnetAZ }}
      CidrBlock: {{ $v.PublicSubnetCIDR }}
      MapPublicIpOnLaunch: {{ $v.PublicSubnetMapPublicIPOnLaunch }}
      Tags:
      - Key: Name
        Value: {{ $v.PublicSubnetName }}
      VpcId: !Ref VPC

  PublicSubnetRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref PublicRouteTable
      SubnetId: !Ref PublicSubnet

  PrivateSubnet:
    Type: AWS::EC2::Subnet
    Properties:
      AvailabilityZone: {{ $v.PrivateSubnetAZ }}
      CidrBlock: {{ $v.PrivateSubnetCIDR }}
      MapPublicIpOnLaunch: {{ $v.PrivateSubnetMapPublicIPOnLaunch }}
      Tags:
      - Key: Name
        Value: {{ $v.PrivateSubnetName }}
      VpcId: !Ref VPC

  PrivateSubnetRouteTableAssociation:
    Type: AWS::EC2::SubnetRouteTableAssociation
    Properties:
      RouteTableId: !Ref PrivateRouteTable
      SubnetId: !Ref PrivateSubnet
{{ end }}`

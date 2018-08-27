package guest

const NatGateway = `{{define "nat_gateway"}}
  {{- $v := .Guest.NATGateway }}
  NATGateway:
    Type: AWS::EC2::NatGateway
    Properties:
      AllocationId:
        Fn::GetAtt:
        - NATEIP
        - AllocationId
      SubnetId: !Ref PublicSubnet
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}
  NATEIP:
    Type: AWS::EC2::EIP
    Properties:
      Domain: vpc
  NATRoute:
    Type: AWS::EC2::Route
    Properties:
      RouteTableId: !Ref PrivateRouteTable
      DestinationCidrBlock: 0.0.0.0/0
      NatGatewayId:
        Ref: "NATGateway"
{{end}}`

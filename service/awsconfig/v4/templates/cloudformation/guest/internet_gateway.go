package guest

const InternetGateway = `{{define "internet_gateway"}}
  InternetGateway:
    Type: AWS::EC2::InternetGateway
    Properties:
      Tags:
        - Key: Name
          Value: {{ .ClusterID }}

  VPCGatewayAttachment:
    Type: AWS::EC2::VPCGatewayAttachment
    DependsOn:
      - PublicRouteTable
      - PrivateRouteTable
    Properties:
      InternetGatewayId:
        Ref: InternetGateway
      VpcId: !Ref VPC

  InternetGatewayRoute:
    Type: AWS::EC2::Route
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      RouteTableId: !Ref PublicRouteTable
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId:
        Ref: InternetGateway
{{end}}`

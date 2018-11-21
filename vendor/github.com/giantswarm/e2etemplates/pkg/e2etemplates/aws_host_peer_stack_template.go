package e2etemplates

const awsHostPeerStackTemplate = `
AWSTemplateFormatVersion: 2010-09-09
Description: Host Peer Stack with VPC peering and route tables for testing purposes
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.11.0.0/16
      Tags:
      - Key: Name
        Value: {{ .Stack.Name }}
  PeerRouteTable0:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ .RouteTable0.Name }}
  PeerRouteTable1:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: {{ .RouteTable1.Name }}
Outputs:
  VPCID:
    Description: Accepter VPC ID
    Value: !Ref VPC
`

package e2etemplates

// TODO AWSHostVPCStack is deprecated. See awsHostPeerStackTemplate instead. We
// track the cleanup in a roadmap story.
//
//     https://github.com/giantswarm/giantswarm/pull/2202
//
const AWSHostVPCStack = `AWSTemplateFormatVersion: 2010-09-09
Description: CI Host Stack with Peering VPC and route tables
Resources:
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.11.0.0/16
      Tags:
      - Key: Name
        Value: ${CLUSTER_NAME}
  PeerRouteTable0:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: ${AWS_ROUTE_TABLE_0}
  PeerRouteTable1:
    Type: AWS::EC2::RouteTable
    Properties:
      VpcId: !Ref VPC
      Tags:
      - Key: Name
        Value: ${AWS_ROUTE_TABLE_1}
Outputs:
  VPCID:
    Description: Accepter VPC ID
    Value: !Ref VPC

`

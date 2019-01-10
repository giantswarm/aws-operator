package guest

const VPC = `{{define "vpc"}}
{{- $v := .Guest.VPC }}
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: {{ $v.CidrBlock }}
      EnableDnsSupport: 'true'
      EnableDnsHostnames: 'true'
      Tags:
      - Key: Name
        Value: {{ $v.ClusterID }}
      - Key: Installation
        Value: {{ $v.InstallationName }}
  VPCPeeringConnection:
    Type: 'AWS::EC2::VPCPeeringConnection'
    Properties:
      VpcId: !Ref VPC
      PeerVpcId: {{ $v.PeerVPCID }}
      PeerOwnerId: '{{ $v.HostAccountID }}'
      PeerRoleArn: {{ $v.PeerRoleArn }}
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}
  VPCS3Endpoint:
    Type: 'AWS::EC2::VPCEndpoint'
    Properties:
      VpcId: !Ref VPC
      RouteTableIds:
        {{- range $v.RouteTableNames }}
        - !Ref .ResourceName
        {{- end}}
      Policy: !Ref VPCS3EndpointPolicy
{{end}}`

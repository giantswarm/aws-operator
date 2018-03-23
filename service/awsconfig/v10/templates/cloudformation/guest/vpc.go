package guest

const VPC = `{{define "vpc"}}
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: {{ .CidrBlock }}
      EnableDnsSupport: 'true'
      EnableDnsHostnames: 'true'
      Tags:
      - Key: Name
        Value: {{ .ClusterID }}
      - Key: Installation
        Value: {{ .InstallationName }}
  VPCPeeringConnection:
    Type: 'AWS::EC2::VPCPeeringConnection'
    Properties:
      VpcId: !Ref VPC
      PeerVpcId: {{ .PeerVPCID }}
      PeerOwnerId: '{{ .HostAccountID }}'
      PeerRoleArn: {{ .PeerRoleArn }}
      Tags:
        - Key: Name
          Value: {{ .ClusterID }}
{{end}}`

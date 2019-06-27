package tccp

const VPC = `
{{define "vpc"}}
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
        - !Ref {{ .ResourceName }}
        {{- end}}
      ServiceName: 'com.amazonaws.{{ $v.Region }}.s3'
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Sid: "{{ $v.ClusterID }}-vpc-s3-endpoint-policy-bucket"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ $v.RegionARN }}:s3:::*"
          - Sid: "{{ $v.ClusterID }}-vpc-s3-endpoint-policy-object"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ $v.RegionARN }}:s3:::*/*"
{{end}}
`

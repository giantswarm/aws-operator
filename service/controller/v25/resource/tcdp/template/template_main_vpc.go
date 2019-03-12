package template

const TemplateMainVPC = `
{{define "vpc"}}
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: {{ .VPC.CIDR }}
      EnableDnsSupport: 'true'
      EnableDnsHostnames: 'true'
      Tags:
      - Key: Name
        Value: {{ .VPC.Cluster.ID }}
      - Key: Installation
        Value: {{ .VPC.Installation }}
  VPCPeeringConnection:
    Type: 'AWS::EC2::VPCPeeringConnection'
    Properties:
      VpcId: !Ref VPC
      PeerVpcId: {{ .VPC.PeerVPCID }}
      PeerOwnerId: '{{ .VPC.ControlPlane.AWSAccountID }}'
      PeerRoleArn: {{ .VPC.PeerRole.Arn }}
      Tags:
        - Key: Name
          Value: {{ .VPC.Cluster.ID }}
  VPCS3Endpoint:
    Type: 'AWS::EC2::VPCEndpoint'
    Properties:
      VpcId: !Ref VPC
      RouteTableIds:
        {{- range .VPC.RouteTableNames }}
        - !Ref {{ .ResourceName }}
        {{- end}}
      ServiceName: 'com.amazonaws.{{ .VPC.Region }}.s3'
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Sid: "{{ .VPC.Cluster.ID }}-vpc-s3-endpoint-policy-bucket"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ .VPC.RegionARN }}:s3:::*"
          - Sid: "{{ .VPC.Cluster.ID }}-vpc-s3-endpoint-policy-object"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ .VPC.RegionARN }}:s3:::*/*"
{{end}}
`

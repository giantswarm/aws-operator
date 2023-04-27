package template

const TemplateMainVPC = `
{{- define "vpc" -}}
{{- $v := .VPC }}
  VPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: {{ $v.CidrBlock }}
      EnableDnsSupport: 'true'
      EnableDnsHostnames: 'true'
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}
  {{- if $v.CIDRBlockAWSCNI }}
  VPCCIDRBlockAWSCNI:
    Type: AWS::EC2::VPCCidrBlock
    DependsOn:
      - VPC
      - VPCPeeringConnection
    Properties:
      CidrBlock: {{ $v.CIDRBlockAWSCNI }}
      VpcId: !Ref VPC
  {{- end }}
  VPCPeeringConnection:
    Type: 'AWS::EC2::VPCPeeringConnection'
    Properties:
      VpcId: !Ref VPC
      PeerVpcId: {{ $v.PeerVPCID }}
      # PeerOwnerId may be a number starting with 0. Cloud Formation is not able
      # to properly deal with that by its own so the configured value must be
      # quoted in order to ensure the peer owner id is properly handled as
      # string. Otherwise stack creation fails.
      PeerOwnerId: "{{ $v.HostAccountID }}"
      PeerRoleArn: {{ $v.PeerRoleArn }}
  VPCS3Endpoint:
    Type: 'AWS::EC2::VPCEndpoint'
    Properties:
      VpcId: !Ref VPC
      RouteTableIds:
        {{- range $v.RouteTableNames }}
        - !Ref {{ .ResourceName }}
        {{- end}}
      ServiceName: com.amazonaws.{{ $v.Region }}.s3
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Sid: "{{ $v.ClusterID }}-vpc-s3-endpoint-policy-bucket"
            Principal: "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ $v.RegionARN }}:s3:::*"
          - Sid: "{{ $v.ClusterID }}-vpc-s3-endpoint-policy-object"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ $v.RegionARN }}:s3:::*/*"
{{- end -}}
`

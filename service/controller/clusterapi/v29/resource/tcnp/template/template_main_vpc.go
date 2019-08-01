package template

const TemplateMainVPC = `
{{- define "vpc" -}}
  VpcCidrBlock:
    Type: AWS::EC2::VPCCidrBlock
    Properties:
      VpcId: {{ .VPC.TCCP.VPC.ID }}
      CidrBlock: {{ .VPC.TCNP.CIDR }}
  VPCS3Endpoint:
    Type: 'AWS::EC2::VPCEndpoint'
    Properties:
      VpcId: {{ .VPC.TCCP.VPC.ID }}
      RouteTableIds:
        {{- range .VPC.RouteTables }}
        - !Ref {{ .Name }}
        {{- end}}
      ServiceName: 'com.amazonaws.{{ .VPC.Region.Name }}.s3'
      PolicyDocument:
        Version: "2012-10-17"
        Statement:
          - Sid: "{{ .VPC.Cluster.ID }}-vpc-s3-endpoint-policy-bucket"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ .VPC.Region.ARN }}:s3:::*"
          - Sid: "{{ .VPC.Cluster.ID }}-vpc-s3-endpoint-policy-object"
            Principal : "*"
            Effect: "Allow"
            Action: "s3:*"
            Resource: "arn:{{ .VPC.Region.ARN }}:s3:::*/*"
{{- end -}}
`

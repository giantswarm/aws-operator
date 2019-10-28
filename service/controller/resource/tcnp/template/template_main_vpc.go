package template

const TemplateMainVPC = `
{{- define "vpc" -}}
  VpcCidrBlock:
    Type: AWS::EC2::VPCCidrBlock
    Properties:
      CidrBlock: {{ .VPC.TCNP.CIDR }}
      VpcId: {{ .VPC.TCCP.VPC.ID }}
  {{- range .VPC.RouteTables }}
  {{ .Route.Name }}:
    Type: AWS::EC2::Route
    Properties:
      DestinationCidrBlock: {{ .ControlPlane.VPC.CIDR }}
      RouteTableId: !Ref {{ .RouteTable.Name }}
      VpcPeeringConnectionId: {{ .TenantCluster.PeeringConnectionID }}
  {{- end }}
  VPCS3Endpoint:
    Type: 'AWS::EC2::VPCEndpoint'
    Properties:
      VpcId: {{ .VPC.TCCP.VPC.ID }}
      RouteTableIds:
        {{- range .VPC.RouteTables }}
        - !Ref {{ .RouteTable.Name }}
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

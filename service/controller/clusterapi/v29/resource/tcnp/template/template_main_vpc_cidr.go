package template

const TemplateMainVPCCIDR = `
{{- define "vpc_cidr" -}}
  VpcCidrBlock:
    Type: AWS::EC2::VPCCidrBlock
    Properties:
      VpcId: {{ .VPCCIDR.TCCP.VPC.ID }}
      CidrBlock: {{ .VPCCIDR.TCNP.CIDR }}
{{- end -}}
`

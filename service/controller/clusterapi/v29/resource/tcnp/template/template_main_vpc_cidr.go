package template

const TemplateMainVPCCIDR = `
{{- define "vpc_cidr" -}}
  VpcCidrBlock:
    Type: AWS::EC2::VPCCidrBlock
    Properties:
      VpcId: {{ .TCCP.VPC.ID }}
      CidrBlock: {{ .TCNP.CIDR }}
{{- end -}}
`

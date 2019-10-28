package template

const TemplateMainInternetGateway = `
{{- define "internet_gateway" -}}
{{- $v := .Guest.InternetGateway -}}
  InternetGateway:
    Type: AWS::EC2::InternetGateway
  VPCGatewayAttachment:
    Type: AWS::EC2::VPCGatewayAttachment
    DependsOn:
      {{- range $v.InternetGateways }}
      - {{ .RouteTable }}
      {{- end }}
    Properties:
      InternetGatewayId:
        Ref: InternetGateway
      VpcId: !Ref VPC
  {{- range $v.InternetGateways }}
  {{ .InternetGatewayRoute }}:
    Type: AWS::EC2::Route
    DependsOn:
      - VPCGatewayAttachment
    Properties:
      RouteTableId: !Ref {{ .RouteTable }}
      DestinationCidrBlock: 0.0.0.0/0
      GatewayId:
        Ref: InternetGateway
  {{- end}}
{{- end -}}
`

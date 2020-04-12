package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  {{- if .Outputs.Route53Enabled -}}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{ end -}}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
  VPCID:
    Value: !Ref VPC
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
{{- end -}}
`

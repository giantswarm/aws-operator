package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  {{- if .Outputs.Route53Enabled -}}
  HostedZoneID: !Ref HostedZone
  InternalHostedZoneID: !Ref InternalHostedZone
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

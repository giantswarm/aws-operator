package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  {{- if .Outputs.Route53Enabled -}}
  APIServerPublicLoadBalancer:
    Value: !GetAtt ApiLoadBalancer.DNSName
  HostedZoneID: 
    Value: !Ref HostedZone
  InternalHostedZoneID: 
    Value: !Ref InternalHostedZone
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

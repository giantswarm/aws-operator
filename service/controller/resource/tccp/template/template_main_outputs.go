package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  DockerVolumeResourceName:
    Value: {{ .Outputs.Master.DockerVolume.ResourceName }}
  {{- if .Outputs.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{- end }}
  MasterImageID:
    Value: {{ .Outputs.Master.ImageID }}
  MasterInstanceResourceName:
    Value: {{ .Outputs.Master.Instance.ResourceName }}
  MasterInstanceType:
    Value: {{ .Outputs.Master.Instance.Type }}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
  VPCID:
    Value: !Ref VPC
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
{{- end -}}
`

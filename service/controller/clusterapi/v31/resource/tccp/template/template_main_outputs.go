package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  DockerVolumeResourceName:
    Value: {{ .Guest.Outputs.Master.DockerVolume.ResourceName }}
  {{- if .Guest.Outputs.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{- end }}
  IngressTargetGroupsIDs:
    Value: !Join [ ',', !Ref {{ .Guest.Outputs.IngressInsecureTargetGroupResourceName }}, !Ref {{ .Guest.Outputs.IngressSecureTargetGroupResourceName }} ]
  MasterImageID:
    Value: {{ .Guest.Outputs.Master.ImageID }}
  MasterInstanceResourceName:
    Value: {{ .Guest.Outputs.Master.Instance.ResourceName }}
  MasterInstanceType:
    Value: {{ .Guest.Outputs.Master.Instance.Type }}
  OperatorVersion:
    Value: {{ .Guest.Outputs.OperatorVersion }}
  VPCID:
    Value: !Ref VPC
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
{{- end -}}
`

package tccp

const Outputs = `
{{- define "outputs" -}}
  DockerVolumeResourceName:
    Value: {{ .Guest.Outputs.Master.DockerVolume.ResourceName }}
  {{- if .Guest.Outputs.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{- end }}
  MasterImageID:
    Value: {{ .Guest.Outputs.Master.ImageID }}
  MasterInstanceResourceName:
    Value: {{ .Guest.Outputs.Master.Instance.ResourceName }}
  MasterInstanceType:
    Value: {{ .Guest.Outputs.Master.Instance.Type }}
  VersionBundleVersion:
    Value:
      Ref: VersionBundleVersionParameter
  VPCID:
    Value: !Ref VPC
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
{{- end -}}
`

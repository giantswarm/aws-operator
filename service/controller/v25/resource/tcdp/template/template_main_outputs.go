package template

const TemplateMainOutputs = `
{{define "outputs"}}
  CloudConfigVersion:
    Value: {{ .Outputs.CloudConfig.Version }}
  DockerVolumeSizeGB:
    Value: {{ .Outputs.DockerVolumeSizeGB }}
  {{ if .Outputs.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{ end }}
  {{ .Outputs.ASG.Key }}:
    Value: !Ref {{ .Outputs.ASG.Ref }}
  ImageID:
    Value: {{ .Outputs.ImageID }}
  InstanceType:
    Value: {{ .Outputs.InstanceType }}
  VersionBundleVersion:
    Value:
      Ref: VersionBundleVersionParameter
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
{{end}}
`

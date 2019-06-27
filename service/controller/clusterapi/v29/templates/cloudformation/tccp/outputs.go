package tccp

const Outputs = `
{{define "outputs"}}
  DockerVolumeResourceName:
    Value: {{ .Guest.Outputs.Master.DockerVolume.ResourceName }}
  {{ if .Guest.Outputs.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{ end }}
  MasterImageID:
    Value: {{ .Guest.Outputs.Master.ImageID }}
  MasterInstanceResourceName:
    Value: {{ .Guest.Outputs.Master.Instance.ResourceName }}
  MasterInstanceType:
    Value: {{ .Guest.Outputs.Master.Instance.Type }}
  MasterCloudConfigVersion:
    Value: {{ .Guest.Outputs.Master.CloudConfig.Version }}
  VPCID:
    Value: !Ref VPC
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
  WorkerASGName:
    Value: !Ref {{ .Guest.Outputs.Worker.ASG.Ref }}
  WorkerDockerVolumeSizeGB:
    Value: {{ .Guest.Outputs.Worker.DockerVolumeSizeGB }}
  WorkerImageID:
    Value: {{ .Guest.Outputs.Worker.ImageID }}
  WorkerInstanceType:
    Value: {{ .Guest.Outputs.Worker.InstanceType }}
  WorkerCloudConfigVersion:
    Value: {{ .Guest.Outputs.Worker.CloudConfig.Version }}
  VersionBundleVersion:
    Value:
      Ref: VersionBundleVersionParameter
{{end}}
`

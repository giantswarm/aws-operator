package tccp

const Outputs = `
{{define "outputs"}}
  DockerVolumeResourceName:
    Value: {{ .Guest.Outputs.Master.DockerVolume.ResourceName }}
  {{ if .Guest.Outputs.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{ end }}
  MasterIgnitionHash:
    Value: {{ .Guest.Outputs.Master.Ignition.Hash }}
  MasterImageID:
    Value: {{ .Guest.Outputs.Master.ImageID }}
  MasterInstanceResourceName:
    Value: {{ .Guest.Outputs.Master.Instance.ResourceName }}
  MasterInstanceType:
    Value: {{ .Guest.Outputs.Master.Instance.Type }}
  VPCID:
    Value: !Ref VPC
  VPCPeeringConnectionID:
    Value: !Ref VPCPeeringConnection
  WorkerASGName:
    Value: !Ref {{ .Guest.Outputs.Worker.ASG.Ref }}
  WorkerDockerVolumeSizeGB:
    Value: {{ .Guest.Outputs.Worker.DockerVolumeSizeGB }}
  WorkerIgnitionHash:
    Value: {{ .Guest.Outputs.Worker.Ignition.Hash }}
  WorkerImageID:
    Value: {{ .Guest.Outputs.Worker.ImageID }}
  WorkerInstanceType:
    Value: {{ .Guest.Outputs.Worker.InstanceType }}
{{end}}
`

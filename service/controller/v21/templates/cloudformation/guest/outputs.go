package guest

const Outputs = `{{define "outputs"}}
{{- $v := .Guest.Outputs }}
Outputs:
  DockerVolumeResourceName:
    Value: {{ $v.Master.DockerVolume.ResourceName }}
  {{ if $v.Route53Enabled }}
  HostedZoneNameServers:
    Value: !Join [ ',', !GetAtt 'HostedZone.NameServers' ]
  {{ end }}
  MasterImageID:
    Value: {{ $v.Master.ImageID }}
  MasterInstanceResourceName:
    Value: {{ $v.Master.Instance.ResourceName }}
  MasterInstanceType:
    Value: {{ $v.Master.Instance.Type }}
  MasterCloudConfigVersion:
    Value: {{ $v.Master.CloudConfig.Version }}
  {{ $v.Worker.ASG.Key }}:
    Value: !Ref {{ $v.Worker.ASG.Ref }}
  WorkerCount:
    Value: {{ $v.Worker.Count }}
  WorkerDockerVolumeSizeGB:
    Value: {{ $v.Worker.DockerVolumeSizeGB }}
  WorkerImageID:
    Value: {{ $v.Worker.ImageID }}
  WorkerInstanceType:
    Value: {{ $v.Worker.InstanceType }}
  WorkerCloudConfigVersion:
    Value: {{ $v.Worker.CloudConfig.Version }}
  VersionBundleVersion:
    Value:
      Ref: VersionBundleVersionParameter
{{end}}`

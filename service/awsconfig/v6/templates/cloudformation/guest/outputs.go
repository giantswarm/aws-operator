package guest

const Outputs = `{{define "outputs"}}
Outputs:
  MasterImageID:
    Value: {{ .Outputs.Master.ImageID }}
  MasterInstanceType:
    Value: {{ .Outputs.Master.InstanceType }}
  MasterCloudConfigVersion:
    Value: {{ .Outputs.Master.CloudConfig.Version }}
  WorkerCount:
    Value: {{ .Outputs.Worker.Count }}
  WorkerImageID:
    Value: {{ .Outputs.Worker.ImageID }}
  WorkerInstanceType:
    Value: {{ .Outputs.Worker.InstanceType }}
  WorkerCloudConfigVersion:
    Value: {{ .Outputs.Worker.CloudConfig.Version }}
  VersionBundleVersion:
    Value: {{ .Outputs.VersionBundle.Version }}
{{end}}`

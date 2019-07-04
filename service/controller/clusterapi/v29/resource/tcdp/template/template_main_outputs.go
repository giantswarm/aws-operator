package template

const TemplateMainOutputs = `
{{ define "outputs" }}
  AutoScalingGroupName:
    Value: !Ref NodePoolAutoScalingGroup
  CloudConfigVersion:
    Value: {{ .Outputs.CloudConfig.Version }}
  DockerVolumeSizeGB:
    Value: {{ .Outputs.DockerVolumeSizeGB }}
  InstanceImage:
    Value: {{ .Outputs.Instance.Image }}
  InstanceType:
    Value: {{ .Outputs.Instance.Type }}
  VersionBundleVersion:
    Value: {{ .Outputs.VersionBundle.Version }}
{{ end }}
`

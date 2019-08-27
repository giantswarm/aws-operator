package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  AutoScalingGroupName:
    Value: !Ref NodePoolAutoScalingGroup
  DockerVolumeSizeGB:
    Value: {{ .Outputs.DockerVolumeSizeGB }}
  InstanceImage:
    Value: {{ .Outputs.Instance.Image }}
  InstanceType:
    Value: {{ .Outputs.Instance.Type }}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
{{- end -}}
`

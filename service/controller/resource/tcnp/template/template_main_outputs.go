package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  DockerVolumeSizeGB:
    Value: {{ .Outputs.DockerVolumeSizeGB }}
  InstanceImage:
    Value: {{ .Outputs.Instance.Image }}
  InstanceType:
    Value: {{ .Outputs.Instance.Type }}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
  ReleaseVersion:
    Value: {{ .Outputs.ReleaseVersion }}
{{- end -}}
`

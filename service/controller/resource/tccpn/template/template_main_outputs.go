package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  InstanceType:
    Value: {{ .Outputs.InstanceType }}
  MasterReplicas:
    Value: {{ .Outputs.MasterReplicas }}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
  ReleaseVersion:
    Value: {{ .Outputs.ReleaseVersion }}
{{- end -}}
`

package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  InstanceType:
    Value: {{ .Outputs.InstanceType }}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
{{- end -}}
`

package template

const TemplateMainOutputs = `
{{- define "outputs" -}}
  InstanceType:
    Value: {{ .Outputs.Instance.Type }}
  OperatorVersion:
    Value: {{ .Outputs.OperatorVersion }}
{{- end -}}
`

package template

const TemplateMain = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Finalizer Cloud Formation Stack.
Resources:
  {{ template "route_tables" . }}
{{ end }}
`

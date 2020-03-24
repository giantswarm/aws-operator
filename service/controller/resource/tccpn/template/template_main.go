package template

const TemplateMain = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Control Plane Nodes Cloud Formation Stack.
Outputs:
  {{ template "outputs" . }}
Resources:
  {{ template "auto_scaling_group" . }}
  {{ template "eni" . }}
  {{ template "etcd_volume" . }}
  {{ template "iam_policies" . }}
  {{ template "launch_configuration" . }}
{{ end }}
`

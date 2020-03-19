package template

const TemplateMain = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Control Plane Nodes Cloud Formation Stack.
Outputs:
  {{ template "outputs" . }}
Resources:
  {{ template "auto_scaling_group" . }}
  {{ template "etcd_volume" . }}
  {{ template "iam_policies" . }}
  {{ template "launch_configuration" . }}
  {{ template "route_tables" . }}
  {{ template "security_groups" . }}
  {{ template "subnets" . }}
{{ end }}
`

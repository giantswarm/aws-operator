package template

const TemplateMain = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Node Pool Cloud Formation Stack.
Outputs:
  {{ template "outputs" . }}
Resources:
  {{ template "auto_scaling_group" . }}
  {{ template "iam_policies" . }}
  {{ template "launch_template" . }}
  {{ template "route_tables" . }}
  {{ template "security_groups" . }}
  {{ template "subnets" . }}
  {{ template "vpc" . }}
{{ end }}
`

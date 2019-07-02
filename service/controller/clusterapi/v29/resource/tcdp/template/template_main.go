package template

const TemplateMain = `
{{ define "main" }}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Data Plane Cloud Formation Stack.
Outputs:
  {{ template "outputs" . }}
Resources:
  {{ template "auto_scaling_group" . }}
  {{ template "iam_policies" . }}
  {{ template "launch_configuration" . }}
  {{ template "lifecycle_hooks" . }}
  {{ template "route_table_association" . }}
  {{ template "security_groups" . }}
  {{ template "subnets" . }}
{{ end }}
`

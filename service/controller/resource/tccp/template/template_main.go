package template

const TemplateMain = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Control Plane Cloud Formation Stack.
Outputs:
  {{ template "outputs" . }}
Resources:
  {{ template "instance" . }}
  {{ template "internet_gateway" . }}
  {{ template "load_balancers" . }}
  {{ template "nat_gateway" . }}
  {{ template "record_sets" . }}
  {{ template "route_tables" . }}
  {{ template "security_groups" . }}
  {{ template "subnets" . }}
  {{ template "vpc" .}}
{{- end -}}
`

package tccp

const Main = `
{{- define "main" -}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Control Plane Cloud Formation Stack.
Outputs:
  {{ template "outputs" . }}
Parameters:
  VersionBundleVersionParameter:
    Type: String
    Description: Sets the VersionBundleVersion used to generate the template.
Resources:
  {{ template "iam_policies" . }}
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

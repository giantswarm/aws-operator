package template

const TemplateMain = `
{{define "main"}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Data Plane Cloud Formation Stack.
Resources:
  {{template "iam_roles" .}}
{{end}}
`

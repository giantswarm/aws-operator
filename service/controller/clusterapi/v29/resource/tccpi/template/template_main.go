package template

const TemplateMain = `
{{define "main"}}
AWSTemplateFormatVersion: 2010-09-09
Description: Tenant Cluster Control Plane Initializer Cloud Formation Stack.
Resources:
  {{template "iam_roles" .}}
{{end}}
`

package hostpre

const Main = `{{define "main"}}AWSTemplateFormatVersion: 2010-09-09
Description: Main Host Pre-Guest CloudFormation stack.
Resources:
  {{template "iam_roles" .}}
{{end}}`

package hostpost

const Main = `{{define "main"}}AWSTemplateFormatVersion: 2010-09-09
Description: Main Host Post-Guest CloudFormation stack.
Resources:
  {{template "record_sets" .}}
  {{template "route_tables" .}}
{{end}}`

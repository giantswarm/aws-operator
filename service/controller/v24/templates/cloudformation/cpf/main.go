package cpf

const Main = `{{define "main"}}AWSTemplateFormatVersion: 2010-09-09
Description: Control Plane Finalizer Cloud Formation Stack.
Resources:
  {{template "record_sets" .}}
  {{template "route_tables" .}}
{{end}}`

package guest

const Main = `{{define "main"}}AWSTemplateFormatVersion: 2010-09-09
Description: Main Guest CloudFormation stack.
Parameters:
  VersionBundleVersionParameter:
    Type: String
    Description: Sets the VersionBundleVersion used to generate the template. 
Resources:
  {{template "vpc" .}}
  {{template "iam_policies" .}}
  {{template "security_groups" .}}
  {{template "route_tables" .}}
  {{template "subnets" .}}
  {{template "internet_gateway" .}}
  {{template "nat_gateway" .}}
  {{template "instance" .}}
  {{template "load_balancers" .}}
  {{template "launch_configuration" .}}
  {{template "lifecycle_hooks" .}}
  {{template "autoscaling_group" .}}
  {{template "record_sets" .}}
{{template "outputs" .}}
{{end}}`

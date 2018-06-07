package hostpost

const Main = `{{define "main"}}AWSTemplateFormatVersion: 2010-09-09
Description: Main Host Post-Guest CloudFormation stack.
Resources:
  GuestNSRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneName: '{{ .BaseDomain }}.'
      Name: '{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
      Type: 'NS'
      TTL: '900'
      ResourceRecords: !Split [ ',', '{{ .GuestHostedZoneNameServers }}' ]
  {{template "record_sets" .}}
  {{template "route_tables" .}}
{{end}}`

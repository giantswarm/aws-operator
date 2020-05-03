package template

const TemplateMainRecordSets = `
{{- define "record_sets" -}}
{{- $v := .RecordSets -}}
{{- if $v.Route53Enabled -}}
{{ range $r := $v.Records }}
  {{ $r.Resource }}:
    Type: AWS::Route53::RecordSet
    Properties:
      ResourceRecords:
      - !Get {{ $r.ENI.Resource }}.PrimaryPrivateIpAddress
      Name: '{{ $r.Value }}.{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      HostedZoneId: {{ $v.HostedZoneID }}
      Type: A
{{- end -}}
{{- end -}}
{{- end -}}
`

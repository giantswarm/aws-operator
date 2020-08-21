package template

const TemplateMainRecordSets = `
{{- define "record_sets" -}}
{{- if .RecordSets.Route53Enabled -}}
  GuestNSRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneID: '{{ .RecordSets.ControlPlanePublicHostedZoneID }}'
      Name: '{{ .RecordSets.ClusterID }}.k8s.{{ .RecordSets.BaseDomain }}.'
      Type: 'NS'
      TTL: '300'
      ResourceRecords: !Split [ ',', '{{ .RecordSets.GuestHostedZoneNameServers }}' ]
{{- end -}}
{{- end -}}
`

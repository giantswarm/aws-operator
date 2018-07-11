package hostpost

const RecordSets = `{{define "record_sets"}}
{{ if .Route53Enabled }}
  GuestNSRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneName: '{{ .BaseDomain }}.'
      Name: '{{ .ClusterID }}.k8s.{{ .BaseDomain }}.'
      Type: 'NS'
      TTL: '900'
      ResourceRecords: !Split [ ',', '{{ .GuestHostedZoneNameServers }}' ]
{{end}}
{{end}}`

package cpf

const RecordSets = `{{ define "record_sets" }}
{{ $v := .RecordSets }}
{{ if $v.Route53Enabled }}
  GuestNSRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneName: '{{ $v.BaseDomain }}.'
      Name: '{{ $v.ClusterID }}.k8s.{{ $v.BaseDomain }}.'
      Type: 'NS'
      TTL: '900'
      ResourceRecords: !Split [ ',', '{{ $v.GuestHostedZoneNameServers }}' ]
{{ end }}
{{ end }}`

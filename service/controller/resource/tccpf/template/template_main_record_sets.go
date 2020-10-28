package template

const TemplateMainRecordSets = `
{{- define "record_sets" -}}
{{- if .RecordSets.Route53Enabled -}}
  GuestNSRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneId: '{{ .RecordSets.ControlPlaneHostedZoneID }}'
      Name: '{{ .RecordSets.ClusterID }}.k8s.{{ .RecordSets.BaseDomain }}.'
      Type: 'NS'
      TTL: '300'
      ResourceRecords: !Split [ ',', '{{ .RecordSets.TenantHostedZoneNameServers }}' ]
  {{ if ne .RecordSets.ControlPlaneInternalHostedZoneID "" }}
  TenantAPIServerRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneId: '{{ .RecordSets.ControlPlaneInternalHostedZoneID }}'
      Name: 'api.{{ .RecordSets.ClusterID }}.k8s.{{ .RecordSets.BaseDomain }}.'
      Type: 'CNAME'
      TTL: '300'
      ResourceRecords:
        - '{{.RecordSets.TenantAPIPublicLoadBalancer }}'
  TenantIngressRecordSet:
    Type: 'AWS::Route53::RecordSet'
    Properties:
      HostedZoneId: '{{ .RecordSets.ControlPlaneInternalHostedZoneID }}'
      Name: '*.{{ .RecordSets.ClusterID }}.k8s.{{ .RecordSets.BaseDomain }}.'
      Type: 'CNAME'
      TTL: '300'
      ResourceRecords:
        - 'ingress.{{ .RecordSets.ClusterID }}.k8s.{{ .RecordSets.BaseDomain }}'
  {{ end }}
{{- end -}}
{{- end -}}
`

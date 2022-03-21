package template

const TemplateMainEtcdVolume = `
{{- define "etcd_volume" -}}
{{ range .EtcdVolume.List }}
  {{ .Resource }}:
    Type: AWS::EC2::Volume
    Properties:
      AvailabilityZone: {{ .AvailabilityZone }}
      Encrypted: true
      Size: 100
      {{- if ne .SnapshotID "" }}
      SnapshotId: {{ .SnapshotID }}
      {{- end }}
      Tags:
      - Key: Name
        Value: {{ .Name }}
      VolumeType: gp3
      {{- if and (ge .Iops 3000) (le .Iops 16000) }}
      Iops: {{ .Iops }}
      {{- end }}
      {{- if and (ge .Throughput 125) (le .Throughput 1000) }}
      Throughput: {{ .Throughput}}
      {{- end }}
{{- end -}}
{{- end -}}
`

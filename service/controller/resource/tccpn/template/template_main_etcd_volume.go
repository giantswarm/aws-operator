package template

const TemplateMainEtcdVolume = `
{{- define "etcd_volume" -}}
  {{- range .EtcdVolume.List }}
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
      VolumeType: gp2
  {{- end }}
{{- end -}}
`

package template

const TemplateMainEtcdVolume = `
{{- define "etcd_volume" -}}
{{- range $v := .EtcdVolume.Volumes }}
  {{ $v.ResourceName }}:
    Type: AWS::EC2::Volume
    Properties:
      AvailabilityZone: {{ $v.AvailabilityZone }}
      Encrypted: true
      Size: 100
      {{- if ne $v.SnapshotID "" }}
      SnapshotId: {{ $v.SnapshotID }}
      {{- end }}
      Tags:
      - Key: Name
        Value: {{ $v.Name }}
      VolumeType: gp2
{{- end -}}
{{- end -}}
`

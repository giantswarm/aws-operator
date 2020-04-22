package template

const TemplateMainEtcdVolume = `
{{- define "etcd_volume" -}}
  EtcdVolume:
    Type: AWS::EC2::Volume
    Properties:
      AvailabilityZone: {{ .EtcdVolume.AvailabilityZone }}
      Encrypted: true
      Size: 100
      SnapshotId: {{ .EtcdVolume.SnapshotID }}
      Tags:
      - Key: Name
        Value: {{ .EtcdVolume.Name }}
      VolumeType: gp2
{{- end -}}
`

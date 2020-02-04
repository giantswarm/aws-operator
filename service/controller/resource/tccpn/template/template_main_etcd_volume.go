package template

const TemplateMainEtcdVolume = `
{{- define "etcd_volume" -}}
  EtcdVolume:
    Type: AWS::EC2::Volume
    Properties:
      AvailabilityZone: {{ $v.EtcdVolume.AvailabilityZone }}
      Encrypted: true
      Size: 100
      SnapshotId: {{ $v.EtcdVolume.SnapshotID }}
      Tags:
      - Key: Name
        Value: {{ $v.EtcdVolume.Name }}
      VolumeType: gp2
{{- end -}}
`

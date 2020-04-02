package template

const TemplateMainInstance = `
{{- define "instance" -}}
{{- $v := .Instance -}}
  EtcdVolume:
    Type: AWS::EC2::Volume
    Properties:
      Encrypted: true
      Size: 100
      VolumeType: gp2
      AvailabilityZone: {{ $v.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ $v.Master.EtcdVolume.Name }}
{{- end -}}
`

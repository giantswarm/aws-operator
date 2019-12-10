package template

const TemplateMainInstance = `
{{- define "instance" -}}
{{- $v := .Instance -}}
  {{ $v.Master.Instance.ResourceName }}:
    Type: "AWS::EC2::Instance"
    Description: Master instance
    DependsOn:
    - {{ $v.Master.DockerVolume.ResourceName }}
    - EtcdVolume
    Properties:
      AvailabilityZone: {{ $v.Master.AZ }}
      DisableApiTermination: true
      IamInstanceProfile: !Ref MasterInstanceProfile
      ImageId: {{ $v.Image.ID }}
      InstanceType: {{ $v.Master.Instance.Type }}
      Monitoring: {{ $v.Master.Instance.Monitoring }}
      SecurityGroupIds:
      - !Ref MasterSecurityGroup
      SubnetId: !Ref {{ $v.Master.PrivateSubnet }}
      UserData:
        Fn::Base64: |
          {
            "ignition": {
          "version": "2.2.0",
          "config": {
            "append": [
              {
                "source": "{{ $v.Master.S3URL }}"
              }
            ]
          }
        },
        "storage": {
          "filesystems": [
            {
              "name": "docker",
              "mount": {
                "device": "/dev/xvdc",
                "wipeFilesystem": true,
                "label": "docker",
                "format": "xfs"
              }
            },
            {
              "name": "log",
              "mount": {
                "device": "/dev/xvdf",
                "wipeFilesystem": true,
                "label": "log",
                "format": "xfs"
              }
            },
            {
              "name": "etcd",
              "mount": {
                "device": "/dev/xvdh",
                "wipeFilesystem": false,
                "label": "etcd",
                "format": "ext4"
              }
            }
          ]
        }
      }
      Tags:
      - Key: Name
        Value: {{ $v.Cluster.ID }}-master
  {{ $v.Master.DockerVolume.ResourceName }}:
    Type: AWS::EC2::Volume
    Properties:
{{ if eq $v.Master.EncrypterBackend "kms" }}
      Encrypted: true
{{ end }}
      Size: 50
      VolumeType: gp2
      AvailabilityZone: {{ $v.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ $v.Master.DockerVolume.Name }}
  EtcdVolume:
    Type: AWS::EC2::Volume
    Properties:
{{ if eq $v.Master.EncrypterBackend "kms" }}
      Encrypted: true
{{ end }}
      Size: 100
      VolumeType: gp2
      AvailabilityZone: {{ $v.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ $v.Master.EtcdVolume.Name }}
  LogVolume:
    Type: AWS::EC2::Volume
    Properties:
{{ if eq $v.Master.EncrypterBackend "kms" }}
      Encrypted: true
{{ end }}
      Size: 100
      VolumeType: gp2
      AvailabilityZone: {{ $v.Master.AZ }}
      Tags:
      - Key: Name
        Value: {{ $v.Master.LogVolume.Name }}
  {{ $v.Master.Instance.ResourceName }}DockerMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ $v.Master.Instance.ResourceName }}
      VolumeId: !Ref {{ $v.Master.DockerVolume.ResourceName }}
      Device: /dev/xvdc
  {{ $v.Master.Instance.ResourceName }}EtcdMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ $v.Master.Instance.ResourceName }}
      VolumeId: !Ref EtcdVolume
      Device: /dev/xvdh
  {{ $v.Master.Instance.ResourceName }}LogMountPoint:
    Type: AWS::EC2::VolumeAttachment
    Properties:
      InstanceId: !Ref {{ $v.Master.Instance.ResourceName }}
      VolumeId: !Ref LogVolume
      Device: /dev/xvdf
{{- end -}}
`

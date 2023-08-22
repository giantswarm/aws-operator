package template

const TemplateMainLaunchTemplate = `
{{- define "launch_template" -}}
  NodePoolLaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: {{ .LaunchTemplate.Name }}
      LaunchTemplateData:
        BlockDeviceMappings:
        - DeviceName: /dev/xvdh
          VirtualName: /dev/xvdh
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Docker.Volume.Size }}
            VolumeType: gp3
        - DeviceName: /dev/xvdg
          VirtualName: /dev/xvdg
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Kubelet.Volume.Size }}
            VolumeType: gp3
        - DeviceName: /dev/xvdf
          VirtualName: /dev/xvdf
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Logging.Volume.Size }}
            VolumeType: gp3
        - DeviceName: /dev/xvdi
          VirtualName: /dev/xvdi
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Containerd.Volume.Size }}
            VolumeType: gp3
        IamInstanceProfile:
          Name: !Ref NodePoolInstanceProfile
        ImageId: {{ .LaunchTemplate.Instance.Image }}
        InstanceType: {{ .LaunchTemplate.Instance.Type }}
        MetadataOptions:
          HttpTokens: {{ .LaunchTemplate.Metadata.HttpTokens }}
          HttpPutResponseHopLimit: 2
        Monitoring:
          Enabled: {{ .LaunchTemplate.Instance.Monitoring }}
        NetworkInterfaces:
          - AssociatePublicIpAddress: false
            DeviceIndex: 0
            Groups:
              - !Ref GeneralSecurityGroup
        TagSpecifications:
        - ResourceType: instance
          Tags:
            - Key: giantswarm.io/release
              Value: {{ .LaunchTemplate.ReleaseVersion }}
        UserData:
          Fn::Base64: |
            {
              "ignition": {
                "version": "3.0.0",
                "config": {
                  "merge": [
                    {
                      "source": "{{ .LaunchTemplate.SmallCloudConfig.S3URL }}"
                    }
                  ]
                }
              },
              "storage": {
                "filesystems": [
                  {
                    "path": "/var/lib/docker",
                    "device": "/dev/xvdc",
                    "wipeFilesystem": true,
                    "label": "docker",
                    "format": "xfs"
                  },
                  {
                    "path": "/var/lib/kubelet",
                    "device": "/dev/xvdg",
                    "wipeFilesystem": true,
                    "label": "kubelet",
                    "format": "xfs"
                  },
                  {
                    "path": "/var/log",
                    "device": "/dev/xvdf",
                    "wipeFilesystem": true,
                    "label": "log",
                    "format": "xfs"
                  },
                  {
                    "path": "/var/lib/containerd",
                    "device": "/dev/xvdi",
                    "wipeFilesystem": true,
                    "label": "containerd",
                    "format": "xfs"
                  }
                ]
              }
            }
{{- end -}}
`

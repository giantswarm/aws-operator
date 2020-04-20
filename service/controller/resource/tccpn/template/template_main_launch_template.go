package template

const TemplateMainLaunchTemplate = `
{{- define "launch_template" -}}
  ControlPlaneNodeLaunchTemplate:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: {{ .LaunchTemplate.ResourceName }}
      LaunchTemplateData:
        BlockDeviceMappings:
        - DeviceName: /dev/xvdc
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Docker.Volume.Size }}
            VolumeType: gp2
        - DeviceName: /dev/xvdg
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Kubelet.Volume.Size }}
            VolumeType: gp2
        - DeviceName: /dev/xvdf
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .LaunchTemplate.BlockDeviceMapping.Logging.Volume.Size }}
            VolumeType: gp2
        IamInstanceProfile:
          Name: !Ref ControlPlaneNodesInstanceProfile
        ImageId: {{ .LaunchTemplate.Instance.Image }}
        InstanceType: {{ .LaunchTemplate.Instance.Type }}
        Monitoring:
          Enabled: {{ .LaunchTemplate.Instance.Monitoring }}
        NetworkInterfaces:
          - AssociatePublicIpAddress: false
            DeviceIndex: 0
            Groups:
            - !Ref {{ .LaunchTemplate.MasterSecurityGroupID }}
        UserData:
          Fn::Base64: |
            {
              "ignition": {
                "version": "2.2.0",
                "config": {
                  "append": [
                    {
                      "source": "{{ .LaunchTemplate.SmallCloudConfig.S3URL }}"
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
                    "name": "kubelet",
                    "mount": {
                      "device": "/dev/xvdg",
                      "wipeFilesystem": true,
                      "label": "kubelet",
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
                  }
                ]
              }
            }
{{- end -}}
`

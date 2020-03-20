package template

const TemplateMainLaunchConfiguration = `
{{- define "launch_configuration" -}}
  ControlPlaneNodeLaunchConfiguration:
    Type: AWS::AutoScaling::LaunchConfiguration
    Properties:
      AssociatePublicIpAddress: false
      BlockDeviceMappings:
      - DeviceName: /dev/xvdc
        Ebs:
          DeleteOnTermination: true
          Encrypted: true
          VolumeSize: {{ .LaunchConfiguration.BlockDeviceMapping.Docker.Volume.Size }}
          VolumeType: gp2
      - DeviceName: /dev/xvdg
        Ebs:
          DeleteOnTermination: true
          Encrypted: true
          VolumeSize: {{ .LaunchConfiguration.BlockDeviceMapping.Kubelet.Volume.Size }}
          VolumeType: gp2
      - DeviceName: /dev/xvdf
        Ebs:
          DeleteOnTermination: true
          Encrypted: true
          VolumeSize: {{ .LaunchConfiguration.BlockDeviceMapping.Logging.Volume.Size }}
          VolumeType: gp2
      IamInstanceProfile: !Ref ControlPlaneNodesInstanceProfile
      ImageId: {{ .LaunchConfiguration.Instance.Image }}
      InstanceType: {{ .LaunchConfiguration.Instance.Type }}
      InstanceMonitoring: {{ .LaunchConfiguration.Instance.Monitoring }}
      SecurityGroups:
      -  {{ .LaunchConfiguration.MasterSecurityGroupID }}
      UserData:
        Fn::Base64: |
          {
            "ignition": {
              "version": "2.2.0",
              "config": {
                "append": [
                  {
                    "source": "{{ .LaunchConfiguration.SmallCloudConfig.S3URL }}"
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

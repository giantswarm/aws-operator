package template

const TemplateMainLaunchTemplate = `
{{- define "launch_template" -}}
{{- $t := .LaunchTemplate -}}
{{- range $i, $cc := $t.SmallCloudConfigs }}
  ControlPlaneNodeLaunchTemplate{{ $i }}:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: {{ $t.ResourceName }}
      LaunchTemplateData:
        BlockDeviceMappings:
        - DeviceName: /dev/xvdc
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ $t.BlockDeviceMapping.Docker.Volume.Size }}
            VolumeType: gp2
        - DeviceName: /dev/xvdg
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ $t.BlockDeviceMapping.Kubelet.Volume.Size }}
            VolumeType: gp2
        - DeviceName: /dev/xvdf
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ $t.BlockDeviceMapping.Logging.Volume.Size }}
            VolumeType: gp2
        IamInstanceProfile:
          Name: !Ref ControlPlaneNodesInstanceProfile
        ImageId: {{ $t.Instance.Image }}
        InstanceType: {{ $t.Instance.Type }}
        Monitoring:
          Enabled: {{ $t.Instance.Monitoring }}
        NetworkInterfaces:
          - AssociatePublicIpAddress: false
            DeviceIndex: 0
            Groups:
            - {{ $t.MasterSecurityGroupID }}
        UserData:
          Fn::Base64: |
            {
              "ignition": {
                "version": "2.2.0",
                "config": {
                  "append": [
                    {
                      "source": "{{ $cc.S3URL }}"
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
{{- end -}}
`

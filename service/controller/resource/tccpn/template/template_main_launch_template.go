package template

const TemplateMainLaunchTemplate = `
{{- define "launch_template" -}}
{{ range .LaunchTemplate.List }}
  {{ .Resource }}:
    Type: AWS::EC2::LaunchTemplate
    Properties:
      LaunchTemplateName: {{ .Name }}
      LaunchTemplateData:
        BlockDeviceMappings:
        - DeviceName: /dev/xvdc
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .BlockDeviceMapping.Docker.Volume.Size }}
            VolumeType: gp2
        - DeviceName: /dev/xvdg
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .BlockDeviceMapping.Kubelet.Volume.Size }}
            VolumeType: gp2
        - DeviceName: /dev/xvdf
          Ebs:
            DeleteOnTermination: true
            Encrypted: true
            VolumeSize: {{ .BlockDeviceMapping.Logging.Volume.Size }}
            VolumeType: gp2
        IamInstanceProfile:
          Name: !Ref ControlPlaneNodesInstanceProfile
        ImageId: {{ .Instance.Image }}
        InstanceType: {{ .Instance.Type }}
        Monitoring:
          Enabled: {{ .Instance.Monitoring }}
        NetworkInterfaces:
          - AssociatePublicIpAddress: false
            DeviceIndex: 0
            Groups:
            - {{ .MasterSecurityGroupID }}
        TagSpecifications:
        - ResourceType: instance
          Tags:
            - Key: giantswarm.io/release
              Value: {{ .ReleaseVersion }}
        UserData:
          Fn::Base64: |
            {
              "ignition": {
                "version": "2.2.0",
                "config": {
                  "append": [
                    {
                      "source": "{{ .SmallCloudConfig.S3URL }}"
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

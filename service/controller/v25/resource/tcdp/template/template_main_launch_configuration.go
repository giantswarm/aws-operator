package template

const TemplateMainLaunchConfiguration = `
{{ define "launch_configuration" }}
  NodePoolLaunchConfiguration:
    Type: AWS::AutoScaling::LaunchConfiguration
    Properties:
      ImageId: {{ .LaunchConfiguration.Instance.Image }}
      SecurityGroups:
      - !Ref NodePoolSecurityGroup
      InstanceType: {{ .LaunchConfiguration.Instance.Type }}
      InstanceMonitoring: {{ .LaunchConfiguration.Instance.Monitoring }}
      IamInstanceProfile: !Ref NodePoolInstanceProfile
      BlockDeviceMappings:

      - DeviceName: /dev/xvdh
        Ebs:
          DeleteOnTermination: true
          VolumeSize: {{ .LaunchConfiguration.BlockDeviceMappings.Docker.Volume.Size }}
          VolumeType: gp2

      - DeviceName: /dev/xvdf
        Ebs:
          DeleteOnTermination: true
          VolumeSize: {{ .LaunchConfiguration.BlockDeviceMappings.Logging.Volume.Size }}
          VolumeType: gp2

      AssociatePublicIpAddress: false
      UserData: {{ .LaunchConfiguration.SmallCloudConfig }}
{{ end }}
`

package template

const TemplateMainLaunchConfiguration = `
{{define "launch_configuration"}}
  NodePoolLaunchConfiguration:
    Type: AWS::AutoScaling::LaunchConfiguration
    Properties:
      ImageId: {{ .LaunchConfiguration.Instance.Image }}
      SecurityGroups:
      - !Ref SecurityGroup
      InstanceType: {{ .LaunchConfiguration.Instance.Type }}
      InstanceMonitoring: {{ .LaunchConfiguration.InstanceMonitoring }}
      IamInstanceProfile: !Ref InstanceProfile
      BlockDeviceMappings:
      {{ range .LaunchConfiguration.BlockDeviceMappings }}
      - DeviceName: "{{ .DeviceName }}"
        Ebs:
          DeleteOnTermination: {{ .DeleteOnTermination }}
          VolumeSize: {{ .Volume.Size }}
          VolumeType: {{ .Volume.Type }}
      {{ end }}
      AssociatePublicIpAddress: {{ .LaunchConfiguration.AssociatePublicIPAddress }}
      UserData: {{ .LaunchConfiguration.SmallCloudConfig }}
{{end}}
`

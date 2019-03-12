package template

const TemplateMainLaunchConfiguration = `
{{define "launch_configuration"}}
  {{ .LaunchConfiguration.ASGType }}LaunchConfiguration:
    Type: "AWS::AutoScaling::LaunchConfiguration"
    Description: {{ .LaunchConfiguration.ASGType }} launch configuration
    Properties:
      ImageId: {{ .LaunchConfiguration.ImageID }}
      SecurityGroups:
      - !Ref SecurityGroup
      InstanceType: {{ .LaunchConfiguration.InstanceType }}
      InstanceMonitoring: {{ .LaunchConfiguration.InstanceMonitoring }}
      IamInstanceProfile: !Ref InstanceProfile
      BlockDeviceMappings:
      {{ range .LaunchConfiguration.BlockDeviceMappings }}
      - DeviceName: "{{ .DeviceName }}"
        Ebs:
          DeleteOnTermination: {{ .DeleteOnTermination }}
          VolumeSize: {{ .VolumeSize }}
          VolumeType: {{ .VolumeType }}
      {{ end }}
      AssociatePublicIpAddress: {{ .LaunchConfiguration.AssociatePublicIPAddress }}
      UserData: {{ .LaunchConfiguration.SmallCloudConfig }}
{{end}}
`

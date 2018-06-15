package guest

const LaunchConfiguration = `{{define "launch_configuration"}}
  {{ .ASGType }}LaunchConfiguration:
    Type: "AWS::AutoScaling::LaunchConfiguration"
    Description: {{ .ASGType }} launch configuration
    Properties:
      KeyName: vault-poc
      ImageId: {{ .WorkerImageID }}
      SecurityGroups:
      - !Ref WorkerSecurityGroup
      InstanceType: {{ .WorkerInstanceType }}
      InstanceMonitoring: {{ .WorkerInstanceMonitoring }}
      IamInstanceProfile: !Ref WorkerInstanceProfile
      BlockDeviceMappings:
      {{ range .WorkerBlockDeviceMappings }}
      - DeviceName: "{{ .DeviceName }}"
        Ebs:
          DeleteOnTermination: {{ .DeleteOnTermination }}
          VolumeSize: {{ .VolumeSize }}
          VolumeType: {{ .VolumeType }}
      {{ end }}
      AssociatePublicIpAddress: {{ .WorkerAssociatePublicIPAddress }}
      UserData: {{ .WorkerSmallCloudConfig }}
{{end}}`

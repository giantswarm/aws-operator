package guest

const LaunchConfiguration = `{{define "launch_configuration"}}
{{- $v := .Guest.LaunchConfiguration }}
  {{ $v.ASGType }}LaunchConfiguration:
    Type: "AWS::AutoScaling::LaunchConfiguration"
    Description: {{ $v.ASGType }} launch configuration
    Properties:
      ImageId: {{ $v.WorkerImageID }}
      SecurityGroups:
      - !Ref WorkerSecurityGroup
      InstanceType: {{ $v.WorkerInstanceType }}
      InstanceMonitoring: {{ $v.WorkerInstanceMonitoring }}
      IamInstanceProfile: !Ref WorkerInstanceProfile
      BlockDeviceMappings:
      {{ range $v.WorkerBlockDeviceMappings }}
      - DeviceName: "{{ .DeviceName }}"
        Ebs:
          DeleteOnTermination: {{ .DeleteOnTermination }}
          VolumeSize: {{ $v.WorkerDockerVolumeSizeGB }}
          VolumeType: {{ .VolumeType }}
      {{ end }}
      AssociatePublicIpAddress: {{ $v.WorkerAssociatePublicIPAddress }}
      UserData: {{ $v.WorkerSmallCloudConfig }}
{{end}}`

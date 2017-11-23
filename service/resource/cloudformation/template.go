package cloudformation

const (
	MainTemplate = `{{define "main"}}AWSTemplateFormatVersion: 2010-09-09
Description: Main CloudFormation stack.
Resources:
  {{template "launch_configuration" .}}
  {{template "autoscaling_group" .}}
{{end}}`

	LaunchConfigurationTemplate = `{{define "launch_configuration"}}{{ .ASGType }}LaunchConfiguration:
    Type: "AWS::AutoScaling::LaunchConfiguration"
    Description: {{ .ASGType }} launch configuration
    Properties:
      ImageId: {{ .ImageID }}
      SecurityGroups:
      - {{ .SecurityGroupID }}
      InstanceType: {{ .InstanceType }}
      IamInstanceProfile: {{ .IAMInstanceProfileName }}
      BlockDeviceMappings:
      {{ range .BlockDeviceMappings }}
      - DeviceName: "{{ .DeviceName }}"
        Ebs:
          DeleteOnTermination: {{ .DeleteOnTermination }}
          VolumeSize: {{ .VolumeSize }}
          VolumeType: {{ .VolumeType }}
      {{ end }}
      AssociatePublicIpAddress: {{ .AssociatePublicIPAddress }}
      UserData: {{ .SmallCloudConfig }}
{{end}}`

	AutoScalingGroupTemplate = `{{define "autoscaling_group"}}  {{ .ASGType }}AutoScalingGroup:
    Type: "AWS::AutoScaling::AutoScalingGroup"
    Properties:
      VPCZoneIdentifier:
        - {{ .SubnetID }}
      AvailabilityZones: [{{ .AZ }}]
      MinSize: {{ .ASGMinSize }}
      MaxSize: {{ .ASGMaxSize }}
      LaunchConfigurationName: !Ref {{ .ASGType }}LaunchConfiguration
      LoadBalancerNames:
        - {{ .LoadBalancerName }}
      HealthCheckGracePeriod: {{ .HealthCheckGracePeriod }}
    UpdatePolicy:
      AutoScalingRollingUpdate:
        # minimum amount of instances that must always be running during a rolling update
        MinInstancesInService: {{ .MinInstancesInService }}
        # only do a rolling update of this amount of instances max
        MaxBatchSize: {{ .MaxBatchSize }}
        # after creating a new instance, pause operations on the ASG for this amount of time
        PauseTime: {{ .RollingUpdatePauseTime }}
{{end}}`
)

package cloudformation

const (
	MainTemplate = `AWSTemplateFormatVersion: 2010-09-09
Description: {{ .ASGType }} autoscaling group
Resources:
  {{ .ASGType }}LaunchConfiguration:
    Type: "AWS::AutoScaling::LaunchConfiguration"
    Description: {{ .ASGType }} launch configuration
    Properties:
      ImageId: !Ref ImageID
      SecurityGroups:
      - !Ref SecurityGroupID
      InstanceType: !Ref InstanceType
      IamInstanceProfile: !Ref IAMInstanceProfileName
      BlockDeviceMappings:
      {{ range .BlockDeviceMappings }}
      - DeviceName: "{{ .DeviceName }}"
        Ebs:
          DeleteOnTermination: {{ .DeleteOnTermination }}
          VolumeSize: {{ .VolumeSize }}
          VolumeType: {{ .VolumeType }}
      {{ end }}
      AssociatePublicIpAddress: !Ref AssociatePublicIPAddress
      UserData: !Ref SmallCloudConfig
      KeyName: !Ref KeyName
  {{ .ASGType }}AutoScalingGroup:
    Type: "AWS::AutoScaling::AutoScalingGroup"
    Properties:
      VPCZoneIdentifier:
        - !Ref SubnetID
      AvailabilityZones:
        - !Ref AZ
      MinSize: !Ref ASGMinSize
      MaxSize: !Ref ASGMaxSize
      LaunchConfigurationName: !Ref {{ .ASGType }}LaunchConfiguration
      LoadBalancerNames:
        - !Ref LoadBalancerName
      HealthCheckGracePeriod: !Ref HealthCheckGracePeriod
    UpdatePolicy:
      AutoScalingRollingUpdate:
        # minimum amount of instances that must always be running during a rolling update
        MinInstancesInService: 2
        # only do a rolling update of this amount of instances max
        MaxBatchSize: 2
        # after creating a new instance, pause operations on the ASG for this amount of time
        PauseTime: PT10S

`
)

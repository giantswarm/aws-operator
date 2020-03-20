package template

const TemplateMainAutoScalingGroup = `
{{- define "auto_scaling_group" -}}
  ControlPlaneNodeAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      VPCZoneIdentifier:
        - {{ .AutoScalingGroup.SubnetID }}
      AvailabilityZones:
        - {{ .AutoScalingGroup.AvailabilityZone }}
      DesiredCapacity: 1
      MinSize: 1
      MaxSize: 1
      LaunchConfigurationName: !Ref ControlPlaneNodeLaunchConfiguration

      # We define a lifecycle hook as part of the ASG in order to drain nodes
      # properly on Node Pool deletion. Earlier we defined a separate lifecycle
      # hook referencing the ASG name. In this setting when deleting a Node Pool
      # the lifecycle hook was never executed. We always want node draining for
      # reliably managing customer workloads.
      LifecycleHookSpecificationList:
        - DefaultResult: CONTINUE
          HeartbeatTimeout: 3600
          LifecycleHookName: ControlPlaneNode
          LifecycleTransition: autoscaling:EC2_INSTANCE_TERMINATING

      # 10 seconds after a new node comes into service, the ASG checks the new
      # instance's health.
      HealthCheckGracePeriod: 10

      MetricsCollection:
        - Granularity: "1Minute"

      Tags:
        - Key: Name
          Value: {{ .AutoScalingGroup.ClusterID }}-master
          PropagateAtLaunch: true
    UpdatePolicy:
      AutoScalingRollingUpdate:

        # Minimum amount of nodes that must always be running during a rolling
        # update.
        MinInstancesInService: 1

        # Maximum amount of nodes being rolled at the same time.
        MaxBatchSize: 1

        # After creating a new instance, pause the rolling update on the ASG for
        # 15 minutes.
        PauseTime: PT15M
{{- end -}}
`

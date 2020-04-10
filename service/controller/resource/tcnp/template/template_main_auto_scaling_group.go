package template

const TemplateMainAutoScalingGroup = `
{{- define "auto_scaling_group" -}}
  NodePoolAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      VPCZoneIdentifier:
      {{- range $s := .AutoScalingGroup.Subnets }}
        - !Ref {{ $s }}
      {{- end }}
      AvailabilityZones:
      {{- range $az := .AutoScalingGroup.AvailabilityZones }}
        - {{ $az }}
      {{- end }}
      DesiredCapacity: {{ .AutoScalingGroup.DesiredCapacity }}
      MinSize: {{ .AutoScalingGroup.MinSize }}
      MaxSize: {{ .AutoScalingGroup.MaxSize }}
      MixedInstancesPolicy:
        LaunchTemplate:
          LaunchTemplateSpecification:
            LaunchTemplateId: !Ref NodePoolLaunchTemplate
            Version: !GetAtt NodePoolLaunchTemplate.LatestVersionNumber
          {{- if  .AutoScalingGroup.LaunchTemplateOverrides }}
          Overrides:
          {{- range $s := .AutoScalingGroup.LaunchTemplateOverrides }}
            - InstanceType: {{ $s.InstanceType }}
              WeightedCapacity: {{ $s.WeightedCapacity }}
          {{- end }}
          {{- end }}
        InstancesDistribution:
          OnDemandBaseCapacity: {{ .AutoScalingGroup.OnDemandBaseCapacity }}
          OnDemandPercentageAboveBaseCapacity: {{ .AutoScalingGroup.OnDemandPercentageAboveBaseCapacity }}
          SpotAllocationStrategy: {{ .AutoScalingGroup.SpotAllocationStrategy }}
          SpotInstancePools: {{ .AutoScalingGroup.SpotInstancePools }}
      # We define a lifecycle hook as part of the ASG in order to drain nodes
      # properly on Node Pool deletion. Earlier we defined a separate lifecycle
      # hook referencing the ASG name. In this setting when deleting a Node Pool
      # the lifecycle hook was never executed. We always want node draining for
      # reliably managing customer workloads.
      LifecycleHookSpecificationList:
        - DefaultResult: CONTINUE
          HeartbeatTimeout: 3600
          LifecycleHookName: {{ .AutoScalingGroup.LifeCycleHookName }}
          LifecycleTransition: autoscaling:EC2_INSTANCE_TERMINATING

      # 10 seconds after a new node comes into service, the ASG checks the new
      # instance's health.
      HealthCheckGracePeriod: 10

      MetricsCollection:
        - Granularity: "1Minute"
      Tags:
        - Key: Name
          Value: {{ .AutoScalingGroup.Cluster.ID }}-worker
          PropagateAtLaunch: true
        - Key: k8s.io/cluster-autoscaler/enabled
          Value: true
          PropagateAtLaunch: false
        - Key: k8s.io/cluster-autoscaler/{{ .AutoScalingGroup.Cluster.ID }}
          Value: true
          PropagateAtLaunch: false
    UpdatePolicy:
      AutoScalingRollingUpdate:

        # Minimum amount of nodes that must always be running during a rolling
        # update.
        MinInstancesInService: {{ .AutoScalingGroup.MinInstancesInService }}

        # Maximum amount of nodes being rolled at the same time.
        MaxBatchSize: {{ .AutoScalingGroup.MaxBatchSize }}

        # After creating a new instance, pause the rolling update on the ASG for
        # 15 minutes.
        PauseTime: PT15M
{{- end -}}
`

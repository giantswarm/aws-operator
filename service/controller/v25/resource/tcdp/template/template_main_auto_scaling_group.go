package template

const TemplateMainAutoScalingGroup = `
{{define "auto_scaling_group"}}
  NodePoolAutoScalingGroup:
    Type: AWS::AutoScaling::AutoScalingGroup
    Properties:
      VPCZoneIdentifier:
      {{- range $s := .AutoScalingGroup.Subnets }}
        - !Ref {{ $s }}
      {{end}}
      AvailabilityZones:
      {{- range $az := .AutoScalingGroup.AvailabilityZones }}
        - {{ $az }}
      {{end}}
      DesiredCapacity: {{ .AutoScalingGroup.DesiredCapacity }}
      MinSize: {{ .AutoScalingGroup.MinSize }}
      MaxSize: {{ .AutoScalingGroup.MaxSize }}
      LaunchConfigurationName: !Ref NodePoolLaunchConfiguration
      LoadBalancerNames:
        - !Ref IngressLoadBalancer

      # 10 seconds after a new node comes into service, the ASG checks the new
      # instance's health.
      HealthCheckGracePeriod: 10

      MetricsCollection:
        - Granularity: "1Minute"
      Tags:
        - Key: Name
          Value: NodePoolAutoScalingGroup
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
{{end}}
`

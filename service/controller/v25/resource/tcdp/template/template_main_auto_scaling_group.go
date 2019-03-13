package template

const TemplateMainAutoScalingGroup = `
{{define "auto_scaling_group"}}
  NodePoolAutoScalingGroup:
    Type: "AWS::AutoScaling::AutoScalingGroup"
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
      HealthCheckGracePeriod: {{ .AutoScalingGroup.HealthCheckGracePeriod }}
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
        # minimum amount of instances that must always be running during a rolling update
        MinInstancesInService: {{ .AutoScalingGroup.MinInstancesInService }}
        # only do a rolling update of this amount of instances max
        MaxBatchSize: {{ .AutoScalingGroup.MaxBatchSize }}
        # after creating a new instance, pause operations on the ASG for this amount of time
        PauseTime: {{ .AutoScalingGroup.RollingUpdatePauseTime }}
{{end}}
`

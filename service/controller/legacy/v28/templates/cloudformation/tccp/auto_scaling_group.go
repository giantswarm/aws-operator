package tccp

const AutoScalingGroup = `
{{define "autoscaling_group"}}
{{- $v := .Guest.AutoScalingGroup }}
  {{ $v.ASGType }}AutoScalingGroup:
    Type: "AWS::AutoScaling::AutoScalingGroup"
    Properties:
      VPCZoneIdentifier:
      {{- range $s := $v.PrivateSubnets }}
        - !Ref {{ $s }}
      {{end}}
      AvailabilityZones:
      {{- range $az := $v.WorkerAZs }}
        - {{ $az }}
      {{end}}
      DesiredCapacity: {{ $v.ASGDesiredCapacity }}
      MinSize: {{ $v.ASGMinSize }}
      MaxSize: {{ $v.ASGMaxSize }}
      LaunchConfigurationName: !Ref {{ $v.ASGType }}LaunchConfiguration
      LoadBalancerNames:
        - !Ref IngressLoadBalancer
      HealthCheckGracePeriod: {{ $v.HealthCheckGracePeriod }}
      MetricsCollection:
        - Granularity: "1Minute"
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-{{ $v.ASGType }}
          PropagateAtLaunch: true
        - Key: k8s.io/cluster-autoscaler/enabled
          Value: true
          PropagateAtLaunch: false
        - Key: k8s.io/cluster-autoscaler/{{ $v.ClusterID }}
          Value: true
          PropagateAtLaunch: false
    UpdatePolicy:
      AutoScalingRollingUpdate:
        # minimum amount of instances that must always be running during a rolling update
        MinInstancesInService: {{ $v.MinInstancesInService }}
        # only do a rolling update of this amount of instances max
        MaxBatchSize: {{ $v.MaxBatchSize }}
        # after creating a new instance, pause operations on the ASG for this amount of time
        PauseTime: {{ $v.RollingUpdatePauseTime }}
{{end}}
`

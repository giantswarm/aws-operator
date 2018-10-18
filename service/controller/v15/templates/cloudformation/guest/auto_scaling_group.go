package guest

const AutoScalingGroup = `{{define "autoscaling_group"}}
{{- $v := .Guest.AutoScalingGroup }}
  {{ $v.ASGType }}AutoScalingGroup:
    Type: "AWS::AutoScaling::AutoScalingGroup"
    Properties:
      AvailabilityZones: [{{ $v.WorkerAZ }}]
      DesiredCapacity: {{ $v.ASGMinSize }}
      HealthCheckGracePeriod: {{ $v.HealthCheckGracePeriod }}
      HealthCheckType: ELB
      MaxSize: {{ $v.ASGMaxSize }}
      MinSize: {{ $v.ASGMinSize }}
      LaunchConfigurationName: !Ref {{ $v.ASGType }}LaunchConfiguration
      LoadBalancerNames:
        - !Ref IngressLoadBalancer
      MetricsCollection:
        - Granularity: "1Minute"
      Tags:
        - Key: Name
          Value: {{ $v.ClusterID }}-{{ $v.ASGType }}
          PropagateAtLaunch: true
      VPCZoneIdentifier:
        - !Ref PrivateSubnet
    UpdatePolicy:
      AutoScalingRollingUpdate:
        # minimum amount of instances that must always be running during a rolling update
        MinInstancesInService: {{ $v.MinInstancesInService }}
        # only do a rolling update of this amount of instances max
        MaxBatchSize: {{ $v.MaxBatchSize }}
        # after creating a new instance, pause operations on the ASG for this amount of time
        PauseTime: {{ $v.RollingUpdatePauseTime }}
{{end}}`

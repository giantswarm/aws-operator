package template

const TemplateMainAutoScalingGroup = `
{{- define "auto_scaling_group" -}}
{{- $HAMasters := .AutoScalingGroup.HAMasters -}}
{{ range .AutoScalingGroup.List }}
  {{ .Resource }}:
    Type: AWS::AutoScaling::AutoScalingGroup
    DependsOn:
    {{- range .DependsOn }}
    - {{ . }}
    {{- end }}
    Properties:
      VPCZoneIdentifier:
        - {{ .SubnetID }}
      AvailabilityZones:
        - {{ .AvailabilityZone }}
      DesiredCapacity: 1
      MinSize: 1
      MaxSize: 1
      MixedInstancesPolicy:
        LaunchTemplate:
          LaunchTemplateSpecification:
            LaunchTemplateId: !Ref {{ .LaunchTemplate.Resource }}
            Version: !GetAtt {{ .LaunchTemplate.Resource }}.LatestVersionNumber
      LoadBalancerNames:
      - {{ .LoadBalancers.ApiInternalName }}
      - {{ .LoadBalancers.ApiName }}
      - {{ .LoadBalancers.EtcdName }}

      {{- if $HAMasters }}
      # We define lifecycle hook only in case of HA masters. In case of 1 masters
      # the draining would not work as the API is down when we try to roll the instance.
      # We define a lifecycle hook as part of the ASG in order to drain nodes
      # properly on Node Pool deletion. Earlier we defined a separate lifecycle
      # hook referencing the ASG name. In this setting when deleting a Node Pool
      # the lifecycle hook was never executed. We always want node draining for
      # reliably managing customer workloads.
      LifecycleHookSpecificationList:
        - DefaultResult: CONTINUE
          HeartbeatTimeout: 3600
          LifecycleHookName: ControlPlane
          LifecycleTransition: autoscaling:EC2_INSTANCE_TERMINATING
      {{- end }}
      # 60 seconds after a new node comes into service, the ASG checks the new
      # instance's health.
      HealthCheckGracePeriod: 60

      MetricsCollection:
        - Granularity: "1Minute"

      Tags:
        - Key: Name
          Value: {{ .ClusterID }}-master
          PropagateAtLaunch: true
    UpdatePolicy:
      AutoScalingRollingUpdate:

        # Minimum amount of nodes that must always be running during a rolling
        # update.
        MinInstancesInService: 0

        # Maximum amount of nodes being rolled at the same time.
        MaxBatchSize: 1

        # We pause the roll of the master ASG for 2 mins to give master
        # time to properly join k8s cluster before rolling another one.
        PauseTime: PT2M
{{- end -}}
{{- end -}}
`

package template

const TemplateMainLifecycleHooks = `
{{- define "lifecycle_hooks" -}}
  NodePoolLifecycleHook:
    Type: AWS::AutoScaling::LifecycleHook
    Properties:
      AutoScalingGroupName:
        Ref: NodePoolAutoScalingGroup
      DefaultResult: CONTINUE
      HeartbeatTimeout: 3600
      LifecycleHookName: NodePool
      LifecycleTransition: autoscaling:EC2_INSTANCE_TERMINATING
{{- end -}}
`

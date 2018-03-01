package guest

const LifecycleHooks = `{{define "lifecycle_hooks"}}
  {{ .LifecycleHooks.LifecycleHook.Name }}LifecycleHook:
    Type: "AWS::AutoScaling::LifecycleHook"
    Properties:
      AutoScalingGroupName:
        Ref: {{ .LifecycleHooks.ASG.Name }}
      DefaultResult: CONTINUE
      HeartbeatTimeout: 300
      LifecycleHookName: {{ .LifecycleHooks.LifecycleHook.Name }}
      LifecycleTransition: "autoscaling:EC2_INSTANCE_TERMINATING"
{{end}}`

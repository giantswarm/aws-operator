package guest

const LifecycleHooks = `{{define "lifecycle_hooks"}}
  {{ .LifecycleHooks.Worker.LifecycleHook.Name }}LifecycleHook:
    Type: "AWS::AutoScaling::LifecycleHook"
    Properties:
      AutoScalingGroupName:
        Ref: {{ .LifecycleHooks.Worker.ASG.Ref }}
      DefaultResult: CONTINUE
      HeartbeatTimeout: 3600
      LifecycleHookName: {{ .LifecycleHooks.Worker.LifecycleHook.Name }}
      LifecycleTransition: "autoscaling:EC2_INSTANCE_TERMINATING"
{{end}}`

package guest

const LifecycleHooks = `{{ define "lifecycle_hooks" }}
{{- $v := .Guest.LifecycleHooks }}
  {{ $v.Worker.LifecycleHook.Name }}LifecycleHook:
    Type: "AWS::AutoScaling::LifecycleHook"
    Properties:
      AutoScalingGroupName:
        Ref: {{ $v.Worker.ASG.Ref }}
      DefaultResult: CONTINUE
      HeartbeatTimeout: 3600
      LifecycleHookName: {{ $v.Worker.LifecycleHook.Name }}
      LifecycleTransition: "autoscaling:EC2_INSTANCE_TERMINATING"
{{ end }}`

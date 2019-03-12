package template

const TemplateMainLifecycleHooks = `
{{ define "lifecycle_hooks" }}
  {{ .LifecycleHooks.Name }}LifecycleHook:
    Type: "AWS::AutoScaling::LifecycleHook"
    Properties:
      AutoScalingGroupName:
        Ref: {{ .LifecycleHooks.ASG.Ref }}
      DefaultResult: CONTINUE
      HeartbeatTimeout: 3600
      LifecycleHookName: {{ .LifecycleHooks.Name }}
      LifecycleTransition: "autoscaling:EC2_INSTANCE_TERMINATING"
{{ end }}
`

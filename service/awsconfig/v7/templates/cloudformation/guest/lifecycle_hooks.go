package guest

const LifecycleHooks = `{{define "lifecycle_hooks"}}
  NodeDrainerLifecycleHook:
    Type: "AWS::AutoScaling::LifecycleHook"
    Properties:
      AutoScalingGroupName:
        Ref: {{ .LifecycleHooks.NodeDrainer.Name }}
      DefaultResult: Continue
      HeartbeatTimeout: 300
      LifecycleHookName: NodeDrainer
      LifecycleTransition: "autoscaling:EC2_INSTANCE_TERMINATING"
{{end}}`

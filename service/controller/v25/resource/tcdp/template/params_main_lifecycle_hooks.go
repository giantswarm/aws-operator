package template

import "github.com/giantswarm/aws-operator/service/controller/v25/key"

type ParamsMainLifecycleHooks struct {
	Worker ParamsMainLifecycleHooksWorker
}

type ParamsMainLifecycleHooksWorker struct {
	ASG           ParamsMainLifecycleHooksASG
	LifecycleHook ParamsMainLifecycleHooksLifecycleHook
}

type ParamsMainLifecycleHooksASG struct {
	Ref string
}

type ParamsMainLifecycleHooksLifecycleHook struct {
	Name string
}

func (a *ParamsMainLifecycleHooks) Adapt(config Config) error {
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.LifecycleHook.Name = key.NodeDrainerLifecycleHookName

	return nil
}

package adapter

import "github.com/giantswarm/aws-operator/service/controller/v15/key"

type lifecycleHooksAdapter struct {
	Worker lifecycleHooksAdapterWorker
}

type lifecycleHooksAdapterWorker struct {
	ASG           lifecycleHooksAdapterASG
	LifecycleHook lifecycleHooksAdapterLifecycleHook
}

type lifecycleHooksAdapterASG struct {
	Ref string
}

type lifecycleHooksAdapterLifecycleHook struct {
	Name string
}

func (a *lifecycleHooksAdapter) Adapt(config Config) error {
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.LifecycleHook.Name = key.NodeDrainerLifecycleHookName

	return nil
}

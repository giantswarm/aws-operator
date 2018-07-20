package adapter

import "github.com/giantswarm/aws-operator/service/controller/v14/key"

type guestLifecycleHooksAdapter struct {
	Worker guestLifecycleHooksAdapterWorker
}

type guestLifecycleHooksAdapterWorker struct {
	ASG           guestLifecycleHooksAdapterASG
	LifecycleHook guestLifecycleHooksAdapterLifecycleHook
}

type guestLifecycleHooksAdapterASG struct {
	Ref string
}

type guestLifecycleHooksAdapterLifecycleHook struct {
	Name string
}

func (a *guestLifecycleHooksAdapter) Adapt(config Config) error {
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.LifecycleHook.Name = key.NodeDrainerLifecycleHookName

	return nil
}

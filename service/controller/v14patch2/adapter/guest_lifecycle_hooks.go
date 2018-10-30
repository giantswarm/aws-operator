package adapter

import "github.com/giantswarm/aws-operator/service/controller/v14patch2/key"

type GuestLifecycleHooksAdapter struct {
	Worker GuestLifecycleHooksAdapterWorker
}

type GuestLifecycleHooksAdapterWorker struct {
	ASG           GuestLifecycleHooksAdapterASG
	LifecycleHook GuestLifecycleHooksAdapterLifecycleHook
}

type GuestLifecycleHooksAdapterASG struct {
	Ref string
}

type GuestLifecycleHooksAdapterLifecycleHook struct {
	Name string
}

func (a *GuestLifecycleHooksAdapter) Adapt(config Config) error {
	a.Worker.ASG.Ref = key.WorkerASGRef
	a.Worker.LifecycleHook.Name = key.NodeDrainerLifecycleHookName

	return nil
}

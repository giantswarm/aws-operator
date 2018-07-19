package adapter

import "github.com/giantswarm/aws-operator/service/controller/v14/key"

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v14/templates/cloudformation/guest/guestLifecycle_hooks.go
//

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

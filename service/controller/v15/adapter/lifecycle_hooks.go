package adapter

import "github.com/giantswarm/aws-operator/service/controller/v15/key"

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/controller/v15/templates/cloudformation/guest/lifecycle_hooks.go
//

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

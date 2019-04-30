package adapter

import "github.com/giantswarm/aws-operator/service/controller/clusterapi/v26/legacykey"

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
	a.Worker.ASG.Ref = legacykey.WorkerASGRef
	a.Worker.LifecycleHook.Name = legacykey.NodeDrainerLifecycleHookName

	return nil
}

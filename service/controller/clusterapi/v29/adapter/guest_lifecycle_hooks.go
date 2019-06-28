package adapter

import (
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

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
	a.Worker.ASG.Ref = key.RefWorkerASG
	a.Worker.LifecycleHook.Name = key.RefNodeDrainer

	return nil
}

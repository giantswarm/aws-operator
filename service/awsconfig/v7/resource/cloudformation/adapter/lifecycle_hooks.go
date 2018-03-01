package adapter

import "github.com/giantswarm/aws-operator/service/awsconfig/v7/key"

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v7/templates/cloudformation/guest/lifecycle_hooks.go
//

type lifecycleHooksAdapter struct {
	ASG           lifecycleHooksAdapterASG
	LifecycleHook lifecycleHooksAdapterLifecycleHook
}

type lifecycleHooksAdapterASG struct {
	Tag string
}

type lifecycleHooksAdapterLifecycleHook struct {
	Name string
}

func (a *lifecycleHooksAdapter) Adapt(config Config) error {
	a.ASG.Tag = key.WorkerASGTag
	a.LifecycleHook.Name = key.NodeDrainerLifecycleHookName

	return nil
}

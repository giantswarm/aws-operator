package adapter

import "github.com/giantswarm/aws-operator/service/awsconfig/v7/key"

// The template related to this adapter can be found in the following import.
//
//     github.com/giantswarm/aws-operator/service/awsconfig/v7/templates/cloudformation/guest/lifecycle_hooks.go
//

type lifecycleHooksAdapter struct {
	NodeDrainer lifecycleHooksAdapterNodeDrainer
}

type lifecycleHooksAdapterNodeDrainer struct {
	Name string
}

func (a *lifecycleHooksAdapter) Adapt(config Config) error {
	a.NodeDrainer.Name = key.WorkerASGName

	return nil
}

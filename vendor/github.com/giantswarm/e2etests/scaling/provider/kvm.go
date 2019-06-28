package provider

import (
	"context"
	"encoding/json"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type KVMConfig struct {
	Clients Clients
	Logger  micrologger.Logger
	Waiter  Waiter

	ClusterID string
}

type KVM struct {
	clients Clients
	logger  micrologger.Logger
	waiter  Waiter

	clusterID string
}

func NewKVM(config KVMConfig) (*KVM, error) {
	if config.Clients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Waiter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Waiter must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	k := &KVM{
		clients: config.Clients,
		logger:  config.Logger,
		waiter:  config.Waiter,

		clusterID: config.ClusterID,
	}

	return k, nil
}

func (k *KVM) AddWorker() error {
	customObject, err := k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	patches := []Patch{
		{
			Op:    "add",
			Path:  "/spec/kvm/workers/-",
			Value: customObject.Spec.KVM.Workers[0],
		},
	}

	b, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Patch(k.clusterID, types.JSONPatchType, b)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *KVM) NumMasters() (int, error) {
	customObject, err := k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.KVM.Masters)

	return num, nil
}

func (k *KVM) NumWorkers() (int, error) {
	customObject, err := k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.KVM.Workers)

	return num, nil
}

func (k *KVM) RemoveWorker() error {
	patches := []Patch{
		{
			Op:   "remove",
			Path: "/spec/kvm/workers/1",
		},
	}

	b, err := json.Marshal(patches)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Patch(k.clusterID, types.JSONPatchType, b)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *KVM) WaitForNodes(ctx context.Context, num int) error {
	err := k.waiter.WaitForNodesReady(ctx, num)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

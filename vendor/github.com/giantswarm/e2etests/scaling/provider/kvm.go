package provider

import (
	"encoding/json"

	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type KVMConfig struct {
	GuestFramework *framework.Guest
	HostFramework  *framework.Host
	Logger         micrologger.Logger

	ClusterID string
}

type KVM struct {
	guestFramework *framework.Guest
	hostFramework  *framework.Host
	logger         micrologger.Logger

	clusterID string
}

func NewKVM(config KVMConfig) (*KVM, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	k := &KVM{
		guestFramework: config.GuestFramework,
		hostFramework:  config.HostFramework,
		logger:         config.Logger,

		clusterID: config.ClusterID,
	}

	return k, nil
}

func (k *KVM) AddWorker() error {
	customObject, err := k.hostFramework.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
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

	_, err = k.hostFramework.G8sClient().ProviderV1alpha1().KVMConfigs("default").Patch(k.clusterID, types.JSONPatchType, b)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *KVM) NumMasters() (int, error) {
	customObject, err := k.hostFramework.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	num := len(customObject.Spec.KVM.Masters)

	return num, nil
}

func (k *KVM) NumWorkers() (int, error) {
	customObject, err := k.hostFramework.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
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

	_, err = k.hostFramework.G8sClient().ProviderV1alpha1().KVMConfigs("default").Patch(k.clusterID, types.JSONPatchType, b)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *KVM) WaitForNodes(num int) error {
	err := k.guestFramework.WaitForNodesUp(num)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

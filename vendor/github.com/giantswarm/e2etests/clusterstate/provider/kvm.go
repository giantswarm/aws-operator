package provider

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KVMConfig struct {
	Clients Clients
	Logger  micrologger.Logger

	ClusterID string
}

type KVM struct {
	clients Clients
	logger  micrologger.Logger

	clusterID string
}

func NewKVM(config KVMConfig) (*KVM, error) {
	if config.Clients == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}

	k := &KVM{
		clients: config.Clients,
		logger:  config.Logger,

		clusterID: config.ClusterID,
	}

	return k, nil
}

func (k *KVM) RebootMaster() error {
	err := k.deleteMasterPod()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *KVM) ReplaceMaster() error {
	err := k.deleteMasterPod()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (k *KVM) deleteMasterPod() error {
	namespace := k.clusterID
	listOptions := metav1.ListOptions{
		LabelSelector: "app=master",
	}

	pods, err := k.clients.K8sClient().CoreV1().Pods(namespace).List(listOptions)
	if err != nil {
		return microerror.Mask(err)
	} else if len(pods.Items) == 0 {
		return microerror.Maskf(notFoundError, "master pod not found")
	} else if len(pods.Items) > 1 {
		return microerror.Maskf(tooManyResultsError, "expected 1 master pod found %d", len(pods.Items))
	}

	masterPod := pods.Items[0]
	err = k.clients.K8sClient().CoreV1().Pods(namespace).Delete(masterPod.Name, &metav1.DeleteOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

package provider

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KVMConfig struct {
	Clients Clients
	Logger  micrologger.Logger

	ClusterID   string
	GithubToken string
}

type KVM struct {
	clients Clients
	logger  micrologger.Logger

	clusterID   string
	githubToken string
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
	if config.GithubToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.GithubToken must not be empty", config)
	}

	k := &KVM{
		clients: config.Clients,
		logger:  config.Logger,

		clusterID:   config.ClusterID,
		githubToken: config.GithubToken,
	}

	return k, nil
}

func (k *KVM) CurrentStatus() (v1alpha1.StatusCluster, error) {
	customObject, err := k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return customObject.Status.Cluster, nil
}

func (k *KVM) CurrentVersion() (string, error) {
	p := &framework.VBVParams{
		Component: "kvm-operator",
		Provider:  "kvm",
		Token:     k.githubToken,
		VType:     "current",
	}
	v, err := framework.GetVersionBundleVersion(p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if v == "" {
		return "", microerror.Mask(versionNotFoundError)
	}

	return v, nil
}

func (k *KVM) NextVersion() (string, error) {
	p := &framework.VBVParams{
		Component: "kvm-operator",
		Provider:  "kvm",
		Token:     k.githubToken,
		VType:     "wip",
	}
	v, err := framework.GetVersionBundleVersion(p)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if v == "" {
		return "", microerror.Mask(versionNotFoundError)
	}

	return v, nil
}

func (k *KVM) UpdateVersion(nextVersion string) error {
	customObject, err := k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Get(k.clusterID, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	customObject.Spec.Cluster.Kubernetes.Kubelet.Labels = ensureLabel(customObject.Spec.Cluster.Kubernetes.Kubelet.Labels, "kvm-operator.giantswarm.io/version", nextVersion)
	customObject.Spec.VersionBundle.Version = nextVersion

	_, err = k.clients.G8sClient().ProviderV1alpha1().KVMConfigs("default").Update(customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

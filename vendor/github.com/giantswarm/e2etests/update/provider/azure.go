package provider

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AzureConfig struct {
	HostFramework *framework.Host
	Logger        micrologger.Logger

	ClusterID   string
	GithubToken string
}

type Azure struct {
	hostFramework *framework.Host
	logger        micrologger.Logger

	clusterID   string
	githubToken string
}

func NewAzure(config AzureConfig) (*Azure, error) {
	if config.HostFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HostFramework must not be empty", config)
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

	a := &Azure{
		hostFramework: config.HostFramework,
		logger:        config.Logger,

		clusterID:   config.ClusterID,
		githubToken: config.GithubToken,
	}

	return a, nil
}

func (a *Azure) CurrentStatus() (v1alpha1.StatusCluster, error) {
	customObject, err := a.hostFramework.G8sClient().ProviderV1alpha1().AzureConfigs("default").Get(a.clusterID, metav1.GetOptions{})
	if err != nil {
		return v1alpha1.StatusCluster{}, microerror.Mask(err)
	}

	return customObject.Status.Cluster, nil
}

func (a *Azure) CurrentVersion() (string, error) {
	p := &framework.VBVParams{
		Component: "azure-operator",
		Provider:  "azure",
		Token:     a.githubToken,
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

func (a *Azure) NextVersion() (string, error) {
	p := &framework.VBVParams{
		Component: "azure-operator",
		Provider:  "azure",
		Token:     a.githubToken,
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

func (a *Azure) UpdateVersion(nextVersion string) error {
	customObject, err := a.hostFramework.G8sClient().ProviderV1alpha1().AzureConfigs("default").Get(a.clusterID, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	customObject.Spec.Cluster.Kubernetes.Kubelet.Labels = ensureLabel(customObject.Spec.Cluster.Kubernetes.Kubelet.Labels, "azure-operator.giantswarm.io/version", nextVersion)
	customObject.Spec.VersionBundle.Version = nextVersion

	_, err = a.hostFramework.G8sClient().ProviderV1alpha1().AzureConfigs("default").Update(customObject)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

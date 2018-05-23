package helmclient

import "k8s.io/helm/pkg/helm"

const (
	tillerDefaultNamespace = "kube-system"
	tillerImageSpec        = "quay.io/giantswarm/tiller:v2.8.2"
	tillerLabelSelector    = "app=helm,name=tiller"
	tillerPodName          = "tiller-giantswarm"
	tillerPort             = 44134
)

// Interface describes the methods provided by the helm client.
type Interface interface {
	// DeleteRelease uninstalls a chart given its release name.
	DeleteRelease(releaseName string, options ...helm.DeleteOption) error
	// EnsureTillerInstalled installs Tiller by creating its deployment and waiting
	// for it to start. A service account and cluster role binding are also created.
	// As a first step, it checks if Tiller is already ready, in which case it
	// returns early.
	EnsureTillerInstalled() error
	// GetReleaseContent gets the current status of the Helm Release. The
	// releaseName is the name of the Helm Release that is set when the Chart
	// is installed.
	GetReleaseContent(releaseName string) (*ReleaseContent, error)
	// GetReleaseHistory gets the current installed version of the Helm Release.
	// The releaseName is the name of the Helm Release that is set when the Helm
	// Chart is installed.
	GetReleaseHistory(releaseName string) (*ReleaseHistory, error)
	// InstallFromTarball installs a Helm Chart packaged in the given tarball.
	InstallFromTarball(path, ns string, options ...helm.InstallOption) error
	// UpdateReleaseFromTarball updates the given release using the chart packaged
	// in the tarball.
	UpdateReleaseFromTarball(releaseName, path string, options ...helm.UpdateOption) error
}

// ReleaseContent returns status information about a Helm Release.
type ReleaseContent struct {
	// Name is the name of the Helm Release.
	Name string
	// Status is the Helm status code of the Release.
	Status string
	// Values are the values provided when installing the Helm Release.
	Values map[string]interface{}
}

// ReleaseHistory returns version information about a Helm Release.
type ReleaseHistory struct {
	// Name is the name of the Helm Release.
	Name string
	// Version is the version of the Helm Chart that has been deployed.
	Version string
}

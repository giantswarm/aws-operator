package helmclient

import (
	"context"

	"k8s.io/helm/pkg/helm"
)

const (
	// defaultMaxHistory is the maximum number of release versions stored per
	// release by default.
	defaultMaxHistory = 10
	// httpClientTimeout is the timeout when pulling tarballs.
	httpClientTimeout = 5
	// runReleaseTestTimeout is the timeout in seconds when running tests.
	runReleaseTestTimout = 300

	defaultTillerImage     = "quay.io/giantswarm/tiller:v2.12.0"
	defaultTillerNamespace = "kube-system"
	roleBindingNamePrefix  = "tiller"
	tillerLabelSelector    = "app=helm,name=tiller"
	tillerPodName          = "tiller-giantswarm"
	tillerPort             = 44134
)

// Interface describes the methods provided by the helm client.
type Interface interface {
	// DeleteRelease uninstalls a chart given its release name.
	DeleteRelease(ctx context.Context, releaseName string, options ...helm.DeleteOption) error
	// EnsureTillerInstalled installs Tiller by creating its deployment and waiting
	// for it to start. A service account and cluster role binding are also created.
	// As a first step, it checks if Tiller is already ready, in which case it
	// returns early.
	EnsureTillerInstalled(ctx context.Context) error
	// EnsureTillerInstalledWithValues installs Tiller by creating its deployment
	// and waiting for it to start. A service account and cluster role binding are
	// also created. Values can be provided to pass through to Tiller
	// and overwrite its deployment defaults.
	EnsureTillerInstalledWithValues(ctx context.Context, values []string) error
	// GetReleaseContent gets the current status of the Helm Release. The
	// releaseName is the name of the Helm Release that is set when the Chart
	// is installed.
	GetReleaseContent(ctx context.Context, releaseName string) (*ReleaseContent, error)
	// GetReleaseHistory gets the current installed version of the Helm Release.
	// The releaseName is the name of the Helm Release that is set when the Helm
	// Chart is installed.
	GetReleaseHistory(ctx context.Context, releaseName string) (*ReleaseHistory, error)
	// InstallReleaseFromTarball installs a Helm Chart packaged in the given tarball.
	InstallReleaseFromTarball(ctx context.Context, path, ns string, options ...helm.InstallOption) error
	// ListReleaseContents gets the current status of all Helm Releases.
	ListReleaseContents(ctx context.Context) ([]*ReleaseContent, error)
	// LoadChart loads a Helm Chart and returns its structure.
	LoadChart(ctx context.Context, chartPath string) (Chart, error)
	// PingTiller proxies the underlying Helm client PingTiller method.
	PingTiller(ctx context.Context) error
	// PullChartTarball downloads a tarball from the provided tarball URL,
	// returning the file path.
	PullChartTarball(ctx context.Context, tarballURL string) (string, error)
	// RunReleaseTest runs the tests for a Helm Release. This is the same
	// action as running the helm test command.
	RunReleaseTest(ctx context.Context, releaseName string, options ...helm.ReleaseTestOption) error
	// UpdateReleaseFromTarball updates the given release using the chart packaged
	// in the tarball.
	UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helm.UpdateOption) error
}

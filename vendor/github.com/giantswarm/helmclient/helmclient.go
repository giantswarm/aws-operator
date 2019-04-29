package helmclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/golang/protobuf/ptypes"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/cmd/helm/installer"
	"k8s.io/helm/pkg/chartutil"
	helmclient "k8s.io/helm/pkg/helm"
	hapichart "k8s.io/helm/pkg/proto/hapi/chart"
	hapirelease "k8s.io/helm/pkg/proto/hapi/release"
	hapiservices "k8s.io/helm/pkg/proto/hapi/services"
)

var (
	helmStatuses = []hapirelease.Status_Code{
		hapirelease.Status_UNKNOWN,
		hapirelease.Status_DEPLOYED,
		hapirelease.Status_DELETED,
		hapirelease.Status_SUPERSEDED,
		hapirelease.Status_FAILED,
		hapirelease.Status_DELETING,
		hapirelease.Status_PENDING_INSTALL,
		hapirelease.Status_PENDING_UPGRADE,
		hapirelease.Status_PENDING_ROLLBACK,
	}
)

// Config represents the configuration used to create a helm client.
type Config struct {
	Fs afero.Fs
	// HelmClient sets a helm client used for all operations of the initiated
	// client. If this is nil, a new helm client will be created for each
	// operation via proper port forwarding. Setting the helm client here manually
	// might only be sufficient for testing or whenever you know what you do.
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	RestConfig      *rest.Config
	TillerImage     string
	TillerNamespace string
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	fs         afero.Fs
	helmClient helmclient.Interface
	httpClient *http.Client
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	restConfig      *rest.Config
	tillerImage     string
	tillerNamespace string
}

// New creates a new configured Helm client.
func New(config Config) (*Client, error) {
	if config.Fs == nil {
		config.Fs = afero.NewOsFs()
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	if config.RestConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RestConfig must not be empty", config)
	}

	if config.TillerImage == "" {
		config.TillerImage = defaultTillerImage
	}
	if config.TillerNamespace == "" {
		config.TillerNamespace = defaultTillerNamespace
	}

	// Set client timeout to prevent leakages.
	httpClient := &http.Client{
		Timeout: time.Second * httpClientTimeout,
	}

	c := &Client{
		fs:         config.Fs,
		helmClient: config.HelmClient,
		httpClient: httpClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		restConfig:      config.RestConfig,
		tillerNamespace: config.TillerNamespace,

		tillerImage: config.TillerImage,
	}

	return c, nil
}

// DeleteRelease uninstalls a chart given its release name.
func (c *Client) DeleteRelease(ctx context.Context, releaseName string, options ...helmclient.DeleteOption) error {
	o := func() error {
		t, err := c.newTunnel()
		if IsTillerNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}
		defer c.closeTunnel(ctx, t)

		_, err = c.newHelmClientFromTunnel(t).DeleteRelease(releaseName, options...)
		if IsReleaseNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewMaxRetries(10, 5*time.Second)
	n := backoff.NewNotifier(c.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// EnsureTillerInstalled installs Tiller by creating its deployment and waiting
// for it to start. A service account and cluster role binding are also created.
// As a first step, it checks if Tiller is already ready, in which case it
// returns early.
func (c *Client) EnsureTillerInstalled(ctx context.Context) error {
	return c.EnsureTillerInstalledWithValues(ctx, []string{})
}

// EnsureTillerInstalledWithValues installs Tiller by creating its deployment
// and waiting for it to start. A service account and cluster role binding are
// also created. As a first step, it checks if Tiller is already ready, in
// which case it returns early. Values can be provided to pass through to Tiller
// and overwrite its deployment.
func (c *Client) EnsureTillerInstalledWithValues(ctx context.Context, values []string) error {
	// Check if Tiller is already present and return early if so.
	{
		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("finding if tiller is installed in namespace %#q", c.tillerNamespace))

		t, err := c.newTunnel()
		defer c.closeTunnel(ctx, t)
		if err != nil {
			// fall through, we may need to create or upgrade Tiller.
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found that tiller is not installed in namespace %#q", c.tillerNamespace))
		} else {
			err = c.newHelmClientFromTunnel(t).PingTiller()
			if err == nil {
				c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("found that tiller is installed in namespace %#q", c.tillerNamespace))
				return nil
			}
		}
	}

	// Create the service account for tiller so it can pull images and do its do.
	{
		name := tillerPodName
		namespace := c.tillerNamespace

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating serviceaccount %#q in namespace %#q", name, namespace))

		serviceAccont := &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}

		_, err := c.k8sClient.CoreV1().ServiceAccounts(namespace).Create(serviceAccont)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("serviceaccount %#q in namespace %#q already exists", name, namespace))
			// fall through
		} else if guest.IsAPINotAvailable(err) {
			return microerror.Maskf(guest.APINotAvailableError, err.Error())
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created serviceaccount %#q in namespace %#q", name, namespace))
		}
	}

	// Create the cluster role binding for tiller so it is allowed to do its job.
	{
		serviceAccountName := tillerPodName
		serviceAccountNamespace := c.tillerNamespace

		name := fmt.Sprintf("%s-%s", roleBindingNamePrefix, serviceAccountNamespace)

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating clusterrolebinding %#q", name))

		i := &rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: serviceAccountNamespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
		}

		_, err := c.k8sClient.RbacV1().ClusterRoleBindings().Create(i)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("clusterrolebinding %#q already exists", name))
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created clusterrolebinding %#q", name))
		}
	}

	var err error
	var installTiller bool
	var pod *corev1.Pod
	var upgradeTiller bool

	{
		o := func() error {
			pod, err = getPod(c.k8sClient, tillerLabelSelector, c.tillerNamespace)
			if IsNotFound(err) {
				// Fall through as we need to install Tiller.
				installTiller = true
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewExponential(1*time.Minute, 5*time.Second)
		n := backoff.NewNotifier(c.logger, context.Background())

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	if !installTiller && pod != nil {
		err = validateTillerVersion(pod, c.tillerImage)
		if IsTillerOutdated(err) {
			upgradeTiller = true
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	i := &installer.Options{
		AutoMountServiceAccountToken: true,
		ImageSpec:                    c.tillerImage,
		MaxHistory:                   defaultMaxHistory,
		Namespace:                    c.tillerNamespace,
		ServiceAccount:               tillerPodName,
		Values:                       values,
	}

	// Install the tiller deployment in the tenant cluster.
	if installTiller && !upgradeTiller {
		err = c.installTiller(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if !installTiller && upgradeTiller {
		err = c.upgradeTiller(ctx, i)
		if err != nil {
			return microerror.Mask(err)
		}
	} else if installTiller && upgradeTiller {
		return microerror.Maskf(executionFailedError, "invalid state cannot both install and upgrade tiller")
	}

	// Wait for tiller to be up and running. When verifying to be able to ping
	// tiller we make sure 3 consecutive pings succeed before assuming everything
	// is fine.
	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for tiller to be up")

		var i int

		o := func() error {
			t, err := c.newTunnel()
			if !installTiller && IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(ctx, t)

			err = c.newHelmClientFromTunnel(t).PingTiller()
			if err != nil {
				i = 0
				return microerror.Mask(err)
			}

			i++
			if i < 3 {
				return microerror.Maskf(executionFailedError, "failed to ping tiller 3 consecutive times")
			}

			return nil
		}
		b := backoff.NewExponential(1*time.Minute, 5*time.Second)
		n := backoff.NewNotifier(c.logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "waited for tiller to be up")
	}

	return nil
}

// GetReleaseContent gets the current status of the Helm Release including any
// values provided when the chart was installed. The releaseName is the name
// of the Helm Release that is set when the Helm Chart is installed.
func (c *Client) GetReleaseContent(ctx context.Context, releaseName string) (*ReleaseContent, error) {
	var err error

	var resp *hapiservices.GetReleaseContentResponse
	{
		o := func() error {
			t, err := c.newTunnel()
			if IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(ctx, t)

			resp, err = c.newHelmClientFromTunnel(t).ReleaseContent(releaseName)
			if IsReleaseNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewMaxRetries(10, 5*time.Second)
		n := backoff.NewNotifier(c.logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	content, err := releaseToReleaseContent(resp.Release)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return content, nil
}

// GetReleaseHistory gets the current installed version of the Helm Release.
// The releaseName is the name of the Helm Release that is set when the Helm
// Chart is installed.
func (c *Client) GetReleaseHistory(ctx context.Context, releaseName string) (*ReleaseHistory, error) {
	var err error
	var resp *hapiservices.GetHistoryResponse
	{
		o := func() error {
			t, err := c.newTunnel()
			if IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(ctx, t)

			resp, err = c.newHelmClientFromTunnel(t).ReleaseHistory(releaseName, helmclient.WithMaxHistory(1))
			if IsReleaseNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewMaxRetries(10, 5*time.Second)

		n := backoff.NewNotifier(c.logger, ctx)

		err = backoff.RetryNotify(o, b, n)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	if len(resp.Releases) > 1 {
		return nil, microerror.Maskf(tooManyResultsError, "%d releases found, expected 1", len(resp.Releases))
	}

	var history *ReleaseHistory
	{
		release := resp.Releases[0]

		var appVersion string
		var version string
		if release.Chart != nil && release.Chart.Metadata != nil {
			appVersion = release.Chart.Metadata.AppVersion
			version = release.Chart.Metadata.Version
		}

		var lastDeployed time.Time
		if release.Info != nil {
			lastDeployed, err = ptypes.Timestamp(release.Info.LastDeployed)
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		history = &ReleaseHistory{
			AppVersion:   appVersion,
			Description:  release.Info.Description,
			LastDeployed: lastDeployed,
			Name:         release.Name,
			Version:      version,
		}
	}

	return history, nil
}

// InstallReleaseFromTarball installs a chart packaged in the given tarball.
func (c *Client) InstallReleaseFromTarball(ctx context.Context, path, ns string, options ...helmclient.InstallOption) error {
	o := func() error {
		t, err := c.newTunnel()
		if IsTillerNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}
		defer c.closeTunnel(ctx, t)

		release, err := c.newHelmClientFromTunnel(t).InstallRelease(path, ns, options...)
		if IsCannotReuseRelease(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if IsReleaseAlreadyExists(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if IsTarballNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if IsYamlConversionFailed(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			if IsInvalidGZipHeader(err) {
				content, readErr := ioutil.ReadFile(path)
				if readErr == nil {
					c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("invalid GZip header, returned release info: %#v, tarball file content %s", release, content), "stack", fmt.Sprintf("%#v", err))
				} else {
					c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("could not read chart tarball %s", path), "stack", fmt.Sprintf("%#v", readErr))
				}
			}
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("err string: %#q", err.Error()))
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewMaxRetries(10, 5*time.Second)
	n := backoff.NewNotifier(c.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// ListReleaseContents gets the current status of all Helm Releases.
func (c *Client) ListReleaseContents(ctx context.Context) ([]*ReleaseContent, error) {
	var releases []*hapirelease.Release
	{
		o := func() error {
			t, err := c.newTunnel()
			if IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(ctx, t)

			next := ""
			for {
				// Note: We explicitly ask for all release statuses,
				// otherwise Helm will only return successfully deployed releases.
				resp, err := c.newHelmClientFromTunnel(t).ListReleases(
					helmclient.ReleaseListStatuses(helmStatuses),
					helmclient.ReleaseListOffset(next),
				)
				if err != nil {
					return microerror.Mask(err)
				}

				releases = append(releases, resp.GetReleases()...)

				next = resp.GetNext()
				if next == "" {
					break
				}
			}

			return nil
		}
		b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
		n := backoff.NewNotifier(c.logger, ctx)

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	// The Helm API considers each version of a release as a separate release,
	// so will return multiple versions of what a sane person would call a 'release'.
	// So, we filter out everything apart from the latest version.
	releases = filterList(releases)

	contents := []*ReleaseContent{}
	for _, release := range releases {
		content, err := releaseToReleaseContent(release)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		contents = append(contents, content)
	}

	return contents, nil
}

// LoadChart loads a Helm Chart and returns relevant parts of its structure.
func (c *Client) LoadChart(ctx context.Context, chartPath string) (Chart, error) {
	helmChart, err := chartutil.Load(chartPath)
	if err != nil {
		return Chart{}, microerror.Mask(err)
	}

	chart, err := newChart(helmChart)
	if err != nil {
		return Chart{}, microerror.Mask(err)
	}

	return chart, nil
}

// PingTiller proxies the underlying Helm client PingTiller method.
func (c *Client) PingTiller(ctx context.Context) error {
	t, err := c.newTunnel()
	if err != nil {
		return microerror.Mask(err)
	}
	defer c.closeTunnel(ctx, t)

	err = c.newHelmClientFromTunnel(t).PingTiller()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// RunReleaseTest runs the tests for a Helm Release. The releaseName is the
// name of the Helm Release that is set when the Helm Chart is installed. This
// is the same action as running the helm test command.
func (c *Client) RunReleaseTest(ctx context.Context, releaseName string, options ...helmclient.ReleaseTestOption) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("running tests for release %#q", releaseName))

	t, err := c.newTunnel()
	if err != nil {
		return microerror.Mask(err)
	}
	defer c.closeTunnel(ctx, t)

	resChan, errChan := c.newHelmClientFromTunnel(t).RunReleaseTest(releaseName, helmclient.ReleaseTestTimeout(int64(runReleaseTestTimout)))
	if IsReleaseNotFound(err) {
		return backoff.Permanent(microerror.Mask(err))
	} else if err != nil {
		return microerror.Mask(err)
	}

	for {
		select {
		case err := <-errChan:
			if err != nil {
				return microerror.Mask(err)
			}
		case res := <-resChan:
			c.logger.LogCtx(ctx, "level", "debug", "message", res.Msg)

			switch res.Status {
			case hapirelease.TestRun_SUCCESS:
				c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ran tests for release %#q", releaseName))
				return nil
			case hapirelease.TestRun_FAILURE:
				return microerror.Maskf(testReleaseFailureError, "failed to run tests for release %#q", releaseName)
			}
		case <-time.After(runReleaseTestTimout * time.Second):
			return microerror.Maskf(testReleaseTimeoutError, "failed to run tests for release %#q", releaseName)
		}
	}
}

// UpdateReleaseFromTarball updates the given release using the chart packaged
// in the tarball.
func (c *Client) UpdateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helmclient.UpdateOption) error {
	o := func() error {
		t, err := c.newTunnel()
		if IsTillerNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}
		defer c.closeTunnel(ctx, t)

		release, err := c.newHelmClientFromTunnel(t).UpdateRelease(releaseName, path, options...)
		if IsReleaseNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if IsYamlConversionFailed(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			if IsInvalidGZipHeader(err) {
				content, readErr := ioutil.ReadFile(path)
				if readErr == nil {
					c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("invalid GZip header, returned release info: %#v, tarball file content %s", release, content), "stack", fmt.Sprintf("%#v", err))
				} else {
					c.logger.LogCtx(ctx, "level", "error", "message", fmt.Sprintf("could not read chart tarball %s", path), "stack", fmt.Sprintf("%#v", readErr))
				}
			}
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("err string: %#q", err.Error()))
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewMaxRetries(10, 5*time.Second)
	n := backoff.NewNotifier(c.logger, ctx)

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) closeTunnel(ctx context.Context, t *k8sportforward.Tunnel) {
	// In case a helm client is configured there is no tunnel and thus we do
	// nothing here.
	if t == nil {
		return
	}

	err := t.Close()
	if err != nil {
		c.logger.LogCtx(ctx, "level", "error", "message", "failed closing tunnel", "stack", fmt.Sprintf("%#v", err))
	}
}

func (c *Client) installTiller(ctx context.Context, installerOptions *installer.Options) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("creating tiller in namespace %#q", c.tillerNamespace))

	o := func() error {
		err := installer.Install(c.k8sClient, installerOptions)
		if errors.IsAlreadyExists(err) {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("tiller in namespace %#q already exists", c.tillerNamespace))
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		} else {
			c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("created tiller in namespace %#q", c.tillerNamespace))
		}

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 5*time.Second)
	n := backoff.NewNotifier(c.logger, context.Background())

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func newChart(helmChart *hapichart.Chart) (Chart, error) {
	if helmChart == nil || helmChart.Metadata == nil {
		return Chart{}, microerror.Maskf(executionFailedError, "expected non nil argument but got %#v", helmChart)
	}

	chart := Chart{
		Version: helmChart.Metadata.Version,
	}

	return chart, nil
}

func (c *Client) newHelmClientFromTunnel(t *k8sportforward.Tunnel) helmclient.Interface {
	// In case a helm client is configured we just go with it.
	if c.helmClient != nil {
		return c.helmClient
	}

	return helmclient.NewClient(
		helmclient.Host(t.LocalAddress()),
		helmclient.ConnectTimeout(5),
	)
}

func (c *Client) newTunnel() (*k8sportforward.Tunnel, error) {
	// In case a helm client is configured we do not need to create any port
	// forwarding.
	if c.helmClient != nil {
		return nil, nil
	}

	pod, err := getPod(c.k8sClient, tillerLabelSelector, c.tillerNamespace)
	if IsNotFound(err) {
		return nil, microerror.Maskf(tillerNotFoundError, "label selector: %#q namespace: %#q", tillerLabelSelector, c.tillerNamespace)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// Do not create a tunnel if tiller is outdated.
	err = validateTillerVersion(pod, c.tillerImage)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var forwarder *k8sportforward.Forwarder
	{
		c := k8sportforward.ForwarderConfig{
			RestConfig: c.restConfig,
		}

		forwarder, err = k8sportforward.NewForwarder(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var tunnel *k8sportforward.Tunnel
	{
		tunnel, err = forwarder.ForwardPort(c.tillerNamespace, pod.Name, tillerPort)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return tunnel, nil
}

func (c *Client) upgradeTiller(ctx context.Context, installerOptions *installer.Options) error {
	c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("upgrading tiller in namespace %#q", c.tillerNamespace))

	o := func() error {
		err := installer.Upgrade(c.k8sClient, installerOptions)
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("upgraded tiller in namespace %#q", c.tillerNamespace))

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 5*time.Second)
	n := backoff.NewNotifier(c.logger, context.Background())

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// filterList returns a list scrubbed of old releases.
// See https://github.com/helm/helm/blob/3a8a797eab0e1d02456c7944bf41631546ee2e47/cmd/helm/list.go#L197.
func filterList(rels []*hapirelease.Release) []*hapirelease.Release {
	idx := map[string]int32{}

	for _, r := range rels {
		name, version := r.GetName(), r.GetVersion()
		if max, ok := idx[name]; ok {
			// check if we have a greater version already
			if max > version {
				continue
			}
		}
		idx[name] = version
	}

	uniq := make([]*hapirelease.Release, 0, len(idx))
	for _, r := range rels {
		if idx[r.GetName()] == r.GetVersion() {
			uniq = append(uniq, r)
		}
	}
	return uniq
}

func getPod(client kubernetes.Interface, labelSelector, namespace string) (*corev1.Pod, error) {
	o := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pods, err := client.CoreV1().Pods(namespace).List(o)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(pods.Items) > 1 {
		return nil, microerror.Maskf(tooManyResultsError, "%d", len(pods.Items))
	}
	if len(pods.Items) == 0 {
		return nil, microerror.Maskf(notFoundError, "%s", labelSelector)
	}

	return &pods.Items[0], nil
}

func getPodImage(pod *corev1.Pod) (string, error) {
	if len(pod.Spec.Containers) > 1 {
		return "", microerror.Maskf(tooManyResultsError, "%d", len(pod.Spec.Containers))
	}
	if len(pod.Spec.Containers) == 0 {
		return "", microerror.Mask(notFoundError)
	}

	tillerImage := pod.Spec.Containers[0].Image
	if tillerImage == "" {
		return "", microerror.Maskf(executionFailedError, "tiller image is empty")
	}

	return tillerImage, nil
}

func validateTillerVersion(pod *corev1.Pod, desiredImage string) error {
	currentImage, err := getPodImage(pod)
	if err != nil {
		return microerror.Mask(err)
	}

	currentVersion, err := parseTillerVersion(currentImage)
	if err != nil {
		return microerror.Mask(err)
	}

	desiredVersion, err := parseTillerVersion(desiredImage)
	if err != nil {
		return microerror.Mask(err)
	}

	if currentVersion.GreaterThan(desiredVersion) {
		return microerror.Maskf(executionFailedError, "current tiller version %#q is greater than desired tiller version %#q", currentVersion.String(), desiredVersion.String())
	}

	if currentVersion.LessThan(desiredVersion) {
		return microerror.Maskf(tillerOutdatedError, "current tiller version %#q is lower than desired tiller version %#q", currentVersion.String(), desiredVersion.String())
	}

	return nil
}

func parseTillerVersion(tillerImage string) (*semver.Version, error) {
	defaultVersion, err := semver.NewVersion("0.0.0")
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Tiller image tag has the version.
	imageParts := strings.Split(tillerImage, ":")
	if len(imageParts) == 1 {
		// No image tag so we upgrade to set the correct version.
		return defaultVersion, nil
	} else if len(imageParts) != 2 {
		return nil, microerror.Maskf(executionFailedError, "tiller image %#q is invalid", tillerImage)
	}

	tag := imageParts[1]
	if tag == "latest" {
		// Uses latest tag so we upgrade to set the correct version.
		return defaultVersion, nil
	}

	version, err := semver.NewVersion(tag)
	if err != nil {
		return nil, microerror.Maskf(executionFailedError, "parsing version %#q failed with error %#q", tag, err)
	}

	return version, nil
}

func releaseToReleaseContent(release *hapirelease.Release) (*ReleaseContent, error) {
	var err error

	// If parameterizable values were passed at release creation time, raw values
	// are returned by the Tiller API and we convert these to a map. First we need
	// to check if there are values actually passed.
	var values chartutil.Values
	if release.Config != nil {
		raw := []byte(release.Config.Raw)
		values, err = chartutil.ReadValues(raw)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	content := &ReleaseContent{
		Name:   release.Name,
		Status: release.Info.Status.Code.String(),
		Values: values.AsMap(),
	}

	return content, nil
}

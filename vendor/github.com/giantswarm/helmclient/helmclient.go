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
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	EnsureTillerInstalledMaxWait time.Duration
	RestConfig                   *rest.Config
	TillerImage                  string
	TillerNamespace              string
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	fs         afero.Fs
	helmClient helmclient.Interface
	httpClient *http.Client
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	ensureTillerInstalledMaxWait time.Duration
	restConfig                   *rest.Config
	tillerImage                  string
	tillerNamespace              string
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

	if config.EnsureTillerInstalledMaxWait == 0 {
		config.EnsureTillerInstalledMaxWait = defaultEnsureTillerInstalledMaxWait
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

		ensureTillerInstalledMaxWait: config.EnsureTillerInstalledMaxWait,
		restConfig:                   config.RestConfig,
		tillerImage:                  config.TillerImage,
		tillerNamespace:              config.TillerNamespace,
	}

	return c, nil
}

// DeleteRelease uninstalls a chart given its release name.
func (c *Client) DeleteRelease(ctx context.Context, releaseName string, options ...helmclient.DeleteOption) error {
	eventName := "delete_release"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	err := c.deleteRelease(ctx, releaseName, options...)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) deleteRelease(ctx context.Context, releaseName string, options ...helmclient.DeleteOption) error {
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

// GetReleaseContent gets the current status of the Helm Release including any
// values provided when the chart was installed. The releaseName is the name
// of the Helm Release that is set when the Helm Chart is installed.
func (c *Client) GetReleaseContent(ctx context.Context, releaseName string) (*ReleaseContent, error) {
	eventName := "get_release_content"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	releaseContent, err := c.getReleaseContent(ctx, releaseName)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return nil, microerror.Mask(err)
	}

	return releaseContent, nil
}

func (c *Client) getReleaseContent(ctx context.Context, releaseName string) (*ReleaseContent, error) {
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
	eventName := "get_release_history"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	releaseContent, err := c.getReleaseHistory(ctx, releaseName)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return nil, microerror.Mask(err)
	}

	return releaseContent, nil
}

func (c *Client) getReleaseHistory(ctx context.Context, releaseName string) (*ReleaseHistory, error) {
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
	eventName := "install_release_from_tarball"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	err := c.installReleaseFromTarball(ctx, path, ns, options...)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) installReleaseFromTarball(ctx context.Context, path, ns string, options ...helmclient.InstallOption) error {
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
	eventName := "list_release_contents"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	releaseContent, err := c.listReleaseContents(ctx)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return nil, microerror.Mask(err)
	}

	return releaseContent, nil
}

func (c *Client) listReleaseContents(ctx context.Context) ([]*ReleaseContent, error) {
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
	eventName := "load_chart"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	chart, err := c.loadChart(ctx, chartPath)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return Chart{}, microerror.Mask(err)
	}

	return chart, nil
}

func (c *Client) loadChart(ctx context.Context, chartPath string) (Chart, error) {
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
	eventName := "ping_tiller"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	err := c.pingTiller(ctx)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) pingTiller(ctx context.Context) error {
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
	eventName := "run_release_test"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	err := c.runReleaseTest(ctx, releaseName, options...)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) runReleaseTest(ctx context.Context, releaseName string, options ...helmclient.ReleaseTestOption) error {
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
	eventName := "update_release_from_tarball"

	t := prometheus.NewTimer(histogram.WithLabelValues(eventName))
	defer t.ObserveDuration()

	err := c.updateReleaseFromTarball(ctx, releaseName, path, options...)
	if err != nil {
		errorGauge.WithLabelValues(eventName).Inc()
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) updateReleaseFromTarball(ctx context.Context, releaseName, path string, options ...helmclient.UpdateOption) error {
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

	pod, err := getPod(c.k8sClient, c.tillerNamespace)
	if IsNotFound(err) {
		return nil, microerror.Maskf(tillerNotFoundError, "field selector: %#q label selector: %#q namespace: %#q", runningPodFieldSelector, tillerLabelSelector, c.tillerNamespace)
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

func getPod(client kubernetes.Interface, namespace string) (*corev1.Pod, error) {
	o := metav1.ListOptions{
		FieldSelector: runningPodFieldSelector,
		LabelSelector: tillerLabelSelector,
	}
	pods, err := client.CoreV1().Pods(namespace).List(o)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	if len(pods.Items) > 1 {
		return nil, microerror.Maskf(tooManyResultsError, "%d", len(pods.Items))
	}
	if len(pods.Items) == 0 {
		return nil, microerror.Mask(notFoundError)
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

	if !currentVersion.Equal(desiredVersion) {
		return microerror.Maskf(tillerInvalidVersionError, "current tiller version %#q does not match desired tiller version %#q", currentVersion.String(), desiredVersion.String())
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

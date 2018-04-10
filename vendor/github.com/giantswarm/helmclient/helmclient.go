package helmclient

import (
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	helmclient "k8s.io/helm/pkg/helm"
)

const (
	connectionTimeoutSecs = 5
)

// Config represents the configuration used to create a helm client.
type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	RestConfig *rest.Config
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	helmClient helmclient.Interface
	logger     micrologger.Logger
}

// New creates a new configured Helm client.
func New(config Config) (*Client, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	if config.RestConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.RestConfig must not be empty", config)
	}

	host, err := setupConnection(config.K8sClient, config.RestConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	helmClient := helmclient.NewClient(helmclient.Host(host), helmclient.ConnectTimeout(connectionTimeoutSecs))

	fmt.Printf("created helm client\n")
	fmt.Printf("pinging tiller\n")
	operation := func() error {
		return helmClient.PingTiller()
	}
	err = backoff.RetryNotify(operation, newCustomExponentialBackoff(), newNotify())
	if err != nil {
		return nil, microerror.Mask(err)
	}
	fmt.Printf("tiller up\n")

	c := &Client{
		helmClient: helmClient,
		logger:     config.Logger,
	}

	return c, nil
}

func newCustomExponentialBackoff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      60 * time.Second,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	return b
}
func newNotify() func(error, time.Duration) {
	return func(err error, delay time.Duration) {
		log.Printf("retrying pinging tiller")
	}
}

// DeleteRelease uninstalls a chart given its release name.
func (c *Client) DeleteRelease(releaseName string, options ...helmclient.DeleteOption) error {
	_, err := c.helmClient.DeleteRelease(releaseName, options...)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// GetReleaseContent gets the current status of the Helm Release including any
// values provided when the chart was installed. The releaseName is the name
// of the Helm Release that is set when the Helm Chart is installed.
func (c *Client) GetReleaseContent(releaseName string) (*ReleaseContent, error) {
	resp, err := c.helmClient.ReleaseContent(releaseName)
	if IsReleaseNotFound(err) {
		return nil, microerror.Maskf(releaseNotFoundError, releaseName)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	// If parameterizable values were passed at release creation time, raw values
	// are returned by the Tiller API and we convert these to a map. First we need
	// to check if there are values actually passed.
	var values chartutil.Values
	if resp.Release.Config != nil {
		raw := []byte(resp.Release.Config.Raw)
		values, err = chartutil.ReadValues(raw)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	content := &ReleaseContent{
		Name:   resp.Release.Name,
		Status: resp.Release.Info.Status.Code.String(),
		Values: values.AsMap(),
	}

	return content, nil
}

// GetReleaseHistory gets the current installed version of the Helm Release.
// The releaseName is the name of the Helm Release that is set when the Helm
// Chart is installed.
func (c *Client) GetReleaseHistory(releaseName string) (*ReleaseHistory, error) {
	var version string

	resp, err := c.helmClient.ReleaseHistory(releaseName, helmclient.WithMaxHistory(1))
	if IsReleaseNotFound(err) {
		return nil, microerror.Maskf(releaseNotFoundError, releaseName)
	} else if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(resp.Releases) > 1 {
		return nil, microerror.Maskf(tooManyResultsError, "%d releases found, expected 1", len(resp.Releases))
	}

	release := resp.Releases[0]
	if release.Chart != nil && release.Chart.Metadata != nil {
		version = release.Chart.Metadata.Version
	}

	history := &ReleaseHistory{
		Name:    release.Name,
		Version: version,
	}

	return history, nil
}

// InstallFromTarball installs a chart packaged in the given tarball.
func (c *Client) InstallFromTarball(path, ns string, options ...helmclient.InstallOption) error {
	_, err := c.helmClient.InstallRelease(path, ns, options...)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// UpdateReleaseFromTarball updates the given release using the chart packaged
// in the tarball.
func (c *Client) UpdateReleaseFromTarball(releaseName, path string, options ...helmclient.UpdateOption) error {
	_, err := c.helmClient.UpdateRelease(releaseName, path, options...)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func setupConnection(client kubernetes.Interface, config *rest.Config) (string, error) {
	podName, err := getPodName(client, tillerLabelSelector, tillerDefaultNamespace)
	if err != nil {
		return "", microerror.Mask(err)
	}
	fmt.Printf("\n")
	fmt.Printf("podName: %#v\n", podName)
	fmt.Printf("\n")

	t := newTunnel(client.CoreV1().RESTClient(), config, tillerDefaultNamespace, podName, tillerPort)
	err = t.forwardPort()
	if err != nil {
		return "", microerror.Mask(err)
	}

	host := fmt.Sprintf("127.0.0.1:%d", t.Local)

	return host, nil
}

func getPodName(client kubernetes.Interface, labelSelector, namespace string) (string, error) {
	pods, err := client.CoreV1().
		Pods(namespace).
		List(metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		return "", microerror.Mask(err)
	}
	if len(pods.Items) > 1 {
		return "", microerror.Mask(tooManyResultsError)
	}
	if len(pods.Items) == 0 {
		return "", microerror.Mask(notFoundError)
	}
	pod := pods.Items[0]
	return pod.Name, nil
}

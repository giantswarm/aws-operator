package helmclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/errors/guest"
	"github.com/giantswarm/k8sportforward"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/cmd/helm/installer"
	"k8s.io/helm/pkg/chartutil"
	helmclient "k8s.io/helm/pkg/helm"
	hapirelease "k8s.io/helm/pkg/proto/hapi/release"
	hapiservices "k8s.io/helm/pkg/proto/hapi/services"
)

const (
	// runReleaseTestTimeout is the timeout in seconds when running tests.
	runReleaseTestTimout = 300
)

// Config represents the configuration used to create a helm client.
type Config struct {
	// HelmClient sets a helm client used for all operations of the initiated
	// client. If this is nil, a new helm client will be created for each
	// operation via proper port forwarding. Setting the helm client here manually
	// might only be sufficient for testing or whenever you know what you do.
	HelmClient helmclient.Interface
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	RestConfig      *rest.Config
	TillerNamespace string
}

// Client knows how to talk with a Helm Tiller server.
type Client struct {
	helmClient helmclient.Interface
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	restConfig      *rest.Config
	tillerNamespace string
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

	if config.TillerNamespace == "" {
		config.TillerNamespace = tillerDefaultNamespace
	}

	c := &Client{
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		restConfig:      config.RestConfig,
		tillerNamespace: config.TillerNamespace,
	}

	return c, nil
}

// DeleteRelease uninstalls a chart given its release name.
func (c *Client) DeleteRelease(releaseName string, options ...helmclient.DeleteOption) error {
	o := func() error {
		t, err := c.newTunnel()
		if IsTillerNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}
		defer c.closeTunnel(t)

		_, err = c.newHelmClientFromTunnel(t).DeleteRelease(releaseName, options...)
		if IsReleaseNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 60*time.Second)
	n := func(err error, delay time.Duration) {
		c.logger.Log("level", "debug", "message", "failed deleting release", "stack", fmt.Sprintf("%#v", err))
	}

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
func (c *Client) EnsureTillerInstalled() error {
	// Check if Tiller is already present and return early if so.
	{
		t, err := c.newTunnel()
		defer c.closeTunnel(t)
		if err != nil {
			// fall through, we may need to create Tiller.
		} else {
			err = c.newHelmClientFromTunnel(t).PingTiller()
			if err == nil {
				return nil
			}
		}
	}

	// Create the service account for tiller so it can pull images and do its do.
	{
		n := c.tillerNamespace
		i := &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: tillerPodName,
			},
		}

		_, err := c.k8sClient.CoreV1().ServiceAccounts(n).Create(i)
		if errors.IsAlreadyExists(err) {
			c.logger.Log("level", "debug", "message", fmt.Sprintf("serviceaccount %s creation failed", tillerPodName), "stack", fmt.Sprintf("%#v", err))
			// fall through
		} else if guest.IsAPINotAvailable(err) {
			return microerror.Maskf(guest.APINotAvailableError, err.Error())
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	// Create the cluster role binding for tiller so it is allowed to do its job.
	{
		i := &rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac.authorization.k8s.io/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: tillerPodName,
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      tillerPodName,
					Namespace: c.tillerNamespace,
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
			c.logger.Log("level", "debug", "message", fmt.Sprintf("clusterrolebinding %s creation failed", tillerPodName), "stack", fmt.Sprintf("%#v", err))
			// fall through
		} else if err != nil {
			return microerror.Mask(err)
		}
	}

	// Install the tiller deployment in the tenant cluster.
	{
		o := func() error {
			i := &installer.Options{
				ImageSpec:      tillerImageSpec,
				MaxHistory:     defaultMaxHistory,
				Namespace:      c.tillerNamespace,
				ServiceAccount: tillerPodName,
			}

			err := installer.Install(c.k8sClient, i)
			if errors.IsAlreadyExists(err) {
				c.logger.Log("level", "debug", "message", "tiller deployment already exists")
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewExponential(2*time.Minute, 5*time.Second)
		n := backoff.NewNotifier(c.logger, context.Background())

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Wait for tiller to be up and running. When verifying to be able to ping
	// tiller we make sure 3 consecutive pings succeed before assuming everything
	// is fine.
	{
		c.logger.Log("level", "debug", "message", "attempt pinging tiller")

		var i int

		o := func() error {
			t, err := c.newTunnel()
			if IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(t)

			err = c.newHelmClientFromTunnel(t).PingTiller()
			if err != nil {
				i = 0
				return microerror.Mask(err)
			}

			i++
			if i < 3 {
				return microerror.Maskf(executionFailedError, "failed pinging tiller")
			}

			return nil
		}
		b := backoff.NewExponential(2*time.Minute, 5*time.Second)
		n := func(err error, delay time.Duration) {
			c.logger.Log("level", "debug", "message", "failed pinging tiller", "stack", fmt.Sprintf("%#v", err))
		}

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Maskf(tillerInstallationFailedError, err.Error())
		}

		c.logger.Log("level", "debug", "message", "succeeded pinging tiller")
	}

	return nil
}

// GetReleaseContent gets the current status of the Helm Release including any
// values provided when the chart was installed. The releaseName is the name
// of the Helm Release that is set when the Helm Chart is installed.
func (c *Client) GetReleaseContent(releaseName string) (*ReleaseContent, error) {
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
			defer c.closeTunnel(t)

			resp, err = c.newHelmClientFromTunnel(t).ReleaseContent(releaseName)
			if IsReleaseNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewExponential(2*time.Minute, 60*time.Second)
		n := func(err error, delay time.Duration) {
			c.logger.Log("level", "debug", "message", "failed fetching release content")
		}

		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return nil, microerror.Mask(err)
		}
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
	var resp *hapiservices.GetHistoryResponse
	{
		o := func() error {
			t, err := c.newTunnel()
			if IsTillerNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}
			defer c.closeTunnel(t)

			resp, err = c.newHelmClientFromTunnel(t).ReleaseHistory(releaseName, helmclient.WithMaxHistory(1))
			if IsReleaseNotFound(err) {
				return backoff.Permanent(microerror.Mask(err))
			} else if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}
		b := backoff.NewExponential(2*time.Minute, 60*time.Second)
		n := func(err error, delay time.Duration) {
			c.logger.Log("level", "debug", "message", "failed fetching release content", "stack", fmt.Sprintf("%#v", err))
		}

		err := backoff.RetryNotify(o, b, n)
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

		var version string
		if release.Chart != nil && release.Chart.Metadata != nil {
			version = release.Chart.Metadata.Version
		}

		history = &ReleaseHistory{
			Name:    release.Name,
			Version: version,
		}
	}

	return history, nil
}

// InstallFromTarball installs a chart packaged in the given tarball.
func (c *Client) InstallFromTarball(path, ns string, options ...helmclient.InstallOption) error {
	o := func() error {
		t, err := c.newTunnel()
		if IsTillerNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}
		defer c.closeTunnel(t)

		release, err := c.newHelmClientFromTunnel(t).InstallRelease(path, ns, options...)
		if IsCannotReuseRelease(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if IsReleaseAlreadyExists(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if IsTarballNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			if IsInvalidGZipHeader(err) {
				content, readErr := ioutil.ReadFile(path)
				if readErr == nil {
					c.logger.Log("level", "debug", "message", fmt.Sprintf("invalid GZip header, returned release info: %#v, tarball file content %s", release, content), "stack", fmt.Sprintf("%#v", err))
				} else {
					c.logger.Log("level", "debug", "message", fmt.Sprintf("could not read chart tarball %s", path), "stack", fmt.Sprintf("%#v", readErr))
				}
			}
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 60*time.Second)
	n := func(err error, delay time.Duration) {
		c.logger.Log("level", "debug", "message", "failed installing from tarball", "stack", fmt.Sprintf("%#v", err))
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// PingTiller proxies the underlying Helm client PingTiller method.
func (c *Client) PingTiller() error {
	t, err := c.newTunnel()
	if err != nil {
		return microerror.Mask(err)
	}
	defer c.closeTunnel(t)

	err = c.newHelmClientFromTunnel(t).PingTiller()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// RunReleaseTest runs the tests for a Helm Release. The releaseName is the
// name of the Helm Release that is set when the Helm Chart is installed. This
// is the same action as running the helm test command.
func (c *Client) RunReleaseTest(releaseName string, options ...helmclient.ReleaseTestOption) error {
	c.logger.Log("level", "debug", "message", fmt.Sprintf("running release tests for '%s'", releaseName))

	t, err := c.newTunnel()
	if err != nil {
		return microerror.Mask(err)
	}
	defer c.closeTunnel(t)

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
			c.logger.Log("level", "debug", "message", res.Msg)

			switch res.Status {
			case hapirelease.TestRun_SUCCESS:
				c.logger.Log("level", "debug", "message", fmt.Sprintf("successfully ran tests for release '%s'", releaseName))
				return nil
			case hapirelease.TestRun_FAILURE:
				return microerror.Maskf(testReleaseFailureError, "'%s' has failed tests", releaseName)
			}
		case <-time.After(runReleaseTestTimout * time.Second):
			return microerror.Mask(testReleaseTimeoutError)
		}
	}
}

// UpdateReleaseFromTarball updates the given release using the chart packaged
// in the tarball.
func (c *Client) UpdateReleaseFromTarball(releaseName, path string, options ...helmclient.UpdateOption) error {
	o := func() error {
		t, err := c.newTunnel()
		if IsTillerNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			return microerror.Mask(err)
		}
		defer c.closeTunnel(t)

		release, err := c.newHelmClientFromTunnel(t).UpdateRelease(releaseName, path, options...)
		if IsReleaseNotFound(err) {
			return backoff.Permanent(microerror.Mask(err))
		} else if err != nil {
			if IsInvalidGZipHeader(err) {
				content, readErr := ioutil.ReadFile(path)
				if readErr == nil {
					c.logger.Log("level", "debug", "message", fmt.Sprintf("invalid GZip header, returned release info: %#v, tarball file content %s", release, content), "stack", fmt.Sprintf("%#v", err))
				} else {
					c.logger.Log("level", "debug", "message", fmt.Sprintf("could not read chart tarball %s", path), "stack", fmt.Sprintf("%#v", readErr))
				}
			}
			return microerror.Mask(err)
		}

		return nil
	}
	b := backoff.NewExponential(2*time.Minute, 60*time.Second)
	n := func(err error, delay time.Duration) {
		c.logger.Log("level", "debug", "message", "failed updating release from tarball", "stack", fmt.Sprintf("%#v", err))
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c *Client) closeTunnel(t *k8sportforward.Tunnel) {
	// In case a helm client is configured there is no tunnel and thus we do
	// nothing here.
	if t == nil {
		return
	}

	err := t.Close()
	if err != nil {
		c.logger.Log("level", "error", "message", "failed closing tunnel", "stack", fmt.Sprintf("%#v", err))
	}
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

	podName, err := getPodName(c.k8sClient, tillerLabelSelector, c.tillerNamespace)
	if IsNotFound(err) {
		return nil, microerror.Maskf(tillerNotFoundError, "label selector: %#q namespace: %#q", tillerLabelSelector, c.tillerNamespace)
	} else if err != nil {
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
		tunnel, err = forwarder.ForwardPort(c.tillerNamespace, podName, tillerPort)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return tunnel, nil
}

func getPodName(client kubernetes.Interface, labelSelector, namespace string) (string, error) {
	o := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	pods, err := client.CoreV1().Pods(namespace).List(o)
	if err != nil {
		return "", microerror.Mask(err)
	}

	if len(pods.Items) > 1 {
		return "", microerror.Maskf(tooManyResultsError, "%d", len(pods.Items))
	}
	if len(pods.Items) == 0 {
		return "", microerror.Maskf(notFoundError, "%s", labelSelector)
	}
	pod := pods.Items[0]

	return pod.Name, nil
}

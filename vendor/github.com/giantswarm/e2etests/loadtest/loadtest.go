package loadtest

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/helm"
)

type Config struct {
	GuestFramework *framework.Guest
	Logger         micrologger.Logger

	ClusterID    string
	CommonDomain string
}

type LoadTest struct {
	guestFramework *framework.Guest
	logger         micrologger.Logger

	clusterID    string
	commonDomain string
}

func New(config Config) (*LoadTest, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.CommonDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}

	s := &LoadTest{
		guestFramework: config.GuestFramework,
		logger:         config.Logger,

		clusterID:    config.ClusterID,
		commonDomain: config.CommonDomain,
	}

	return s, nil
}

func (l *LoadTest) Test(ctx context.Context) error {
	var err error

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "installing loadtest-app")

		err = l.InstallTestApp(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "installed loadtest-app")
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "waiting for loadtest-app to be ready")

		err = l.CheckTestAppIsInstalled(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "loadtest-app is ready")
	}

	return nil
}

func (l *LoadTest) InstallTestApp(ctx context.Context) error {
	var err error

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: l.logger,

			Address:      CNRAddress,
			Organization: CNROrganization,
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:    l.logger,
			K8sClient: l.guestFramework.K8sClient(),

			RestConfig: l.guestFramework.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.EnsureTillerInstalled(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var loadTestEndpoint string
	{
		loadTestEndpoint = fmt.Sprintf("testapp.%s.%s", l.clusterID, l.commonDomain)

		l.logger.Log("level", "debug", "message", fmt.Sprintf("loadtest-app endpoint is %#q", loadTestEndpoint))
	}

	var jsonValues []byte
	{
		values := LoadTestApp{
			Ingress: LoadTestAppIngress{
				Hosts: []string{
					loadTestEndpoint,
				},
			},
		}

		jsonValues, err = json.Marshal(values)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Install the e2e app chart in the tenant cluster.
	{
		l.logger.Log("level", "debug", "message", "installing loadtest-app for testing")

		tarballPath, err := apprClient.PullChartTarball(ctx, ChartName, ChartChannel)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.InstallReleaseFromTarball(ctx, tarballPath, ChartNamespace, helm.ValueOverrides(jsonValues))
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.Log("level", "debug", "message", "installed loadtest-app for testing")
	}

	return nil
}

func (l *LoadTest) CheckTestAppIsInstalled(ctx context.Context) error {
	var podCount = 1

	l.logger.Log("level", "debug", "message", fmt.Sprintf("waiting for %d pods of the loadtest-app to be up", podCount))

	o := func() error {
		lo := metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=testapp-chart",
		}
		l, err := l.guestFramework.K8sClient().CoreV1().Pods("test").List(lo)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(l.Items) != podCount {
			return microerror.Maskf(waitError, "want %d pods found %d", podCount, len(l.Items))
		}

		return nil
	}

	b := backoff.NewConstant(backoff.ShortMaxWait, backoff.ShortMaxInterval)
	n := func(err error, delay time.Duration) {
		l.logger.Log("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	l.logger.Log("level", "debug", "message", fmt.Sprintf("found %d pods of the e2e-app", podCount))

	return nil
}

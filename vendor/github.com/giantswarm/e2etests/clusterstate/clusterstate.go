package clusterstate

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/e2etests/clusterstate/provider"
)

type Config struct {
	GuestFramework *framework.Guest
	Logger         micrologger.Logger
	Provider       provider.Interface
}

type ClusterState struct {
	guestFramework *framework.Guest
	logger         micrologger.Logger
	provider       provider.Interface
}

func New(config Config) (*ClusterState, error) {
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Provider == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Provider must not be empty", config)
	}

	s := &ClusterState{
		guestFramework: config.GuestFramework,
		logger:         config.Logger,
		provider:       config.Provider,
	}

	return s, nil
}

func (c *ClusterState) Test(ctx context.Context) error {
	var err error

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "installing e2e-app")

		err = c.InstallTestApp()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "installed e2e-app")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "checking test app is installed")

		err = c.CheckTestAppIsInstalled()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "test app is installed")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "rebooting master")

		err = c.provider.RebootMaster()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "rebooted master")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting api to go down")

		err = c.guestFramework.WaitForAPIDown()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "api is down")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = c.guestFramework.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "checking test app is installed")

		err = c.CheckTestAppIsInstalled()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "test app is installed")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "replacing master node")

		err = c.provider.ReplaceMaster()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "master node replaced")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting api to go down")

		err = c.guestFramework.WaitForAPIDown()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "api is down")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "waiting for guest cluster")

		err = c.guestFramework.WaitForGuestReady()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "guest cluster ready")
	}

	{
		c.logger.LogCtx(ctx, "level", "debug", "message", "checking test app is installed")

		err = c.CheckTestAppIsInstalled()
		if err != nil {
			return microerror.Mask(err)
		}

		c.logger.LogCtx(ctx, "level", "debug", "message", "test app is installed")
	}

	return nil
}

func (c *ClusterState) InstallTestApp() error {
	var err error

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: c.logger,

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
			Logger:    c.logger,
			K8sClient: c.guestFramework.K8sClient(),

			RestConfig: c.guestFramework.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.EnsureTillerInstalled()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	// Install the e2e app chart in the guest cluster.
	{
		c.logger.Log("level", "debug", "message", "installing e2e-app for testing")

		tarballPath, err := apprClient.PullChartTarball(ChartName, ChartChannel)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.InstallFromTarball(tarballPath, ChartNamespace)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (c *ClusterState) CheckTestAppIsInstalled() error {
	waitCount := 0

	// Wait for e2e app to be up.
	for {
		c.logger.Log("level", "debug", "message", "waiting for 2 pods of the e2e-app to be up")

		o := metav1.ListOptions{
			LabelSelector: "app=e2e-app",
		}
		l, err := c.guestFramework.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(l.Items) != 2 {
			c.logger.Log("level", "debug", "message", fmt.Sprintf("found %d pods", len(l.Items)))

			time.Sleep(3 * time.Second)
			waitCount++

			continue
		}

		// Stop after 90 seconds if pods not found.
		if waitCount == 30 {
			return microerror.Maskf(notFoundError, "e2e-app pods not found")
		}

		c.logger.Log("level", "debug", "message", "found 2 pods of the e2e-app")

		break
	}

	return nil
}

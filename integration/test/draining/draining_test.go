// +build k8srequired

package scaling

import (
	"fmt"
	"testing"
	"time"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
	ChartChannel    = "stable"
	ChartName       = "e2e-app-chart"
	ChartNamespace  = "default"
)

func Test_Draining(t *testing.T) {
	var err error

	var newLogger micrologger.Logger
	{
		c := micrologger.Config{}

		newLogger, err = micrologger.New(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	var apprClient *apprclient.Client
	{
		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: newLogger,

			Address:      CNRAddress,
			Organization: CNROrganization,
		}

		apprClient, err = apprclient.New(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	var helmClient *helmclient.Client
	{
		c := helmclient.Config{
			Logger:     newLogger,
			K8sClient:  g.K8sClient(),
			RestConfig: g.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	// Install the e2e app chart in the guest cluster.
	{
		tarballPath, err := apprClient.PullChartTarball(ChartName, ChartChannel)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = helmClient.InstallFromTarball(tarballPath, ChartNamespace)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	// wait for e2e app to be up
	for {
		list, err := g.K8sClient().Core().Pods(ChartNamespace).List(metav1.ListOptions{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		fmt.Printf("\n")
		for _, i := range list.Items {
			fmt.Printf("%#v\n", i)
		}
		fmt.Printf("\n")

		time.Sleep(60 * time.Second)
	}

	// continuously request e2e app
	// scale down guest cluster
	// verify no requests where failing
}

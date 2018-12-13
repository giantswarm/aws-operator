// +build k8srequired

package draining

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/service/controller/v21/key"
)

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
	ChartChannel    = "stable"
	ChartName       = "e2e-app-chart"
	ChartNamespace  = "e2e-app"
)

// Test_Draining launches the e2e-app in a guest cluster and requests it
// continuously to verify its availability while scaling down a guest cluster
// node.
//
// TODO bring down the failure tolerance to 0. Right now we accept 10% of failed
// requests against the e2e-app while a node draining happens. This has to be
// investigated further in the future.
//
func Test_Draining(t *testing.T) {
	ctx := context.Background()

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
			Logger:    newLogger,
			K8sClient: config.Guest.K8sClient(),

			RestConfig: config.Guest.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = helmClient.EnsureTillerInstalled(ctx)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	// Install the e2e app chart in the guest cluster.
	{
		newLogger.Log("level", "debug", "message", "installing e2e-app for testing")

		tarballPath, err := apprClient.PullChartTarball(ChartName, ChartChannel)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = helmClient.InstallReleaseFromTarball(ctx, tarballPath, ChartNamespace)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	// wait for e2e app to be up
	for {
		newLogger.Log("level", "debug", "message", "waiting for 2 pods of the e2e-app to be up")

		o := metav1.ListOptions{
			LabelSelector: "app=e2e-app",
		}
		l, err := config.Guest.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		if len(l.Items) != 2 {
			newLogger.Log("level", "debug", "message", fmt.Sprintf("found %d pods", len(l.Items)))
			time.Sleep(3 * time.Second)
			continue
		}

		break
	}

	// continuously request e2e app
	var failure float64
	var success float64
	done := make(chan struct{}, 1)
	{
		newLogger.Log("level", "debug", "message", "continuously requesting e2e-app")

		go func() {
			for {
				time.Sleep(200 * time.Millisecond)

				select {
				case <-done:
					return
				default:
					tlsConfig, err := rest.TLSConfigFor(config.Guest.RestConfig())
					if err != nil {
						fmt.Printf("expected %#v got %#v (%s)\n", nil, err, err.Error())
						continue
					}
					client := &http.Client{
						Transport: &http.Transport{
							TLSClientConfig: tlsConfig,
						},
					}

					restClient := config.Guest.K8sClient().Discovery().RESTClient()
					u := restClient.Get().AbsPath("api", "v1", "namespaces", "e2e-app", "services", "e2e-app:8000", "proxy/").URL()
					resp, err := client.Get(u.String())
					if err != nil {
						fmt.Printf("expected %#v got %#v (%s)\n", nil, err, err.Error())
						continue
					} else {
						defer resp.Body.Close()
					}

					b, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("expected %#v got %#v (%s)\n", nil, err, err.Error())
						continue
					}

					var r E2EAppResponse
					err = json.Unmarshal(b, &r)
					if err != nil {
						fmt.Printf("expected %#v got %#v (%s)\n", nil, err, err.Error())
						continue
					}

					if r.Name != "e2e-app" {
						failure++
					} else if r.Source != "https://github.com/giantswarm/e2e-app" {
						failure++
					} else {
						success++
					}
				}
			}
		}()
	}

	{
		newLogger.Log("level", "debug", "message", "verifying e2e-app availability 10 more seconds")
		time.Sleep(10 * time.Second)

		newLogger.Log("level", "debug", "message", "scaling down guest cluster worker")

		// scale down guest cluster
		masterCount, err := numberOfMasters(env.ClusterID())
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		workerCount, err := numberOfWorkers(env.ClusterID())
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		err = removeWorker(env.ClusterID())
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		err = config.Guest.WaitForNodesUp(masterCount + workerCount - 1)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		newLogger.Log("level", "debug", "message", "verifying e2e-app availability 10 more seconds")
		time.Sleep(10 * time.Second)
		close(done)
	}

	{
		newLogger.Log("level", "debug", "message", "validating test data")

		acceptable := float64(20)
		percOfFail := failure * 100 / success

		newLogger.Log("level", "debug", "message", fmt.Sprintf("failure count is %f", failure))
		newLogger.Log("level", "debug", "message", fmt.Sprintf("success count is %f", success))
		newLogger.Log("level", "debug", "message", fmt.Sprintf("ration is %f of failures", percOfFail))

		if percOfFail > acceptable {
			t.Fatalf("expected %#v got %#v", fmt.Sprintf("less than %f percent of failures", acceptable), fmt.Sprintf("%f of failures", percOfFail))
		}
	}
}

func numberOfMasters(clusterID string) (int, error) {
	cluster, err := config.Host.AWSCluster(clusterID)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.MasterCount(*cluster), nil
}

func numberOfWorkers(clusterID string) (int, error) {
	cluster, err := config.Host.AWSCluster(clusterID)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.WorkerCount(*cluster), nil
}

func removeWorker(clusterID string) error {
	patch := make([]framework.PatchSpec, 1)
	patch[0].Op = "remove"
	patch[0].Path = "/spec/aws/workers/1"

	return config.Host.ApplyAWSConfigPatch(patch, clusterID)
}

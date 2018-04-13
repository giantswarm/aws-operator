// +build k8srequired

package draining

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
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
	"github.com/giantswarm/aws-operator/service/awsconfig/v10/key"
)

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
	ChartChannel    = "stable"
	ChartName       = "e2e-app-chart"
	ChartNamespace  = "e2e-app"
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

	newLogger.Log("level", "debug", "message", "initializing clients")

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
			K8sClient: g.K8sClient(),

			RestConfig: g.RestConfig(),
		}

		helmClient, err = helmclient.New(c)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = helmClient.InstallTiller()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	newLogger.Log("level", "debug", "message", "installing e2e-app for testing")

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

	newLogger.Log("level", "debug", "message", "waiting for 2 pods of the e2e-app to be up")

	// wait for e2e app to be up
	for {
		o := metav1.ListOptions{
			LabelSelector: "app=e2e-app",
		}
		l, err := g.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
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

	newLogger.Log("level", "debug", "message", "continuously requesting e2e-app")

	// continuously request e2e app
	var failure int
	var success int
	done := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				tlsConfig, err := rest.TLSConfigFor(g.RestConfig())
				if err != nil {
					fmt.Printf("expected %#v got %#v", nil, err)
				}
				client := &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: tlsConfig,
					},
				}

				restClient := g.K8sClient().Discovery().RESTClient()
				u := restClient.Get().AbsPath("api", "v1", "proxy", "namespaces", "e2e-app", "services", "e2e-app:8000", "proxy").URL()
				resp, err := client.Get(u.String())
				if err != nil {
					nErr, ok := err.(*net.OpError)
					if ok {
						fmt.Printf("expected %#v got %#v", nil, nErr.Err)
					} else {
						fmt.Printf("expected %#v got %#v", nil, err)
					}
				}
				defer resp.Body.Close()

				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("expected %#v got %#v", nil, err)
				}

				var r E2EAppResponse
				err = json.Unmarshal(b, &r)
				if err != nil {
					fmt.Printf("expected %#v got %#v", nil, err)
				}

				if r.Name != "e2e-app" {
					failure++
				} else if r.Source != "https://github.com/giantswarm/e2e-app" {
					failure++
				} else {
					success++
				}

				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

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
	err = g.WaitForNodesUp(masterCount + workerCount - 1)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	newLogger.Log("level", "debug", "message", "verifying e2e-app availability 10 more seconds")
	time.Sleep(10 * time.Second)
	close(done)

	newLogger.Log("level", "debug", "message", "validating test data")

	newLogger.Log("level", "debug", "message", fmt.Sprintf("failure count is %d", failure))
	newLogger.Log("level", "debug", "message", fmt.Sprintf("success count is %d", success))

	// TODO verify no requests where failing
	if success == 0 {
		t.Fatalf("expected %#v got %#v", "more than 0 successes", "0")
	}
}

func numberOfMasters(clusterID string) (int, error) {
	cluster, err := h.AWSCluster(clusterID)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.MasterCount(*cluster), nil
}

func numberOfWorkers(clusterID string) (int, error) {
	cluster, err := h.AWSCluster(clusterID)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.WorkerCount(*cluster), nil
}

func removeWorker(clusterID string) error {
	patch := make([]framework.PatchSpec, 1)
	patch[0].Op = "remove"
	patch[0].Path = "/spec/aws/workers/1"

	return h.ApplyAWSConfigPatch(patch, clusterID)
}

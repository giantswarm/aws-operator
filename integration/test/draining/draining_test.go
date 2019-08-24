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
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/service/controller/legacy/v29/key"
)

const (
	CNRAddress      = "https://quay.io"
	CNROrganization = "giantswarm"
)

const (
	ChartChannel   = "stable"
	ChartName      = "e2e-app-chart"
	ChartNamespace = "e2e-app"
)

const (
	e2eAppName   = "e2e-app"
	e2eAppSource = "https://github.com/giantswarm/e2e-app"
)

// Test_Draining launches the e2e-app in a tenant cluster and requests it
// continuously to verify its availability while scaling down a tenant cluster
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

	// Install the e2e app chart in the tenant cluster.
	{
		newLogger.Log("level", "debug", "message", "installing e2e-app for testing")

		tarballPath, err := apprClient.PullChartTarball(ctx, ChartName, ChartChannel)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		err = helmClient.InstallReleaseFromTarball(ctx, tarballPath, ChartNamespace)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		newLogger.Log("level", "debug", "message", "installed e2e-app for testing")
	}

	// wait for e2e app to be up
	{
		newLogger.Log("level", "debug", "message", "waiting for 2 pods of the e2e-app to be up")

		o := func() error {
			o := metav1.ListOptions{
				LabelSelector: "app=e2e-app",
			}
			l, err := config.Guest.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
			if err != nil {
				return microerror.Mask(err)
			}

			if len(l.Items) != 2 {
				return microerror.Maskf(podError, "not all pods created")
			}

			for _, p := range l.Items {
				if p.Status.Phase != v1.PodRunning {
					return microerror.Maskf(podError, "%#q is not running", p.GetName())
				}
			}

			return nil
		}
		b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

		err := backoff.Retry(o, b)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		newLogger.Log("level", "debug", "message", "waited for 2 pods of the e2e-app to be up")
	}

	// continuously request e2e app
	var failure float64
	var success float64
	done := make(chan struct{}, 1)
	{
		requestE2EApp := func() error {
			tlsConfig, err := rest.TLSConfigFor(config.Guest.RestConfig())
			if err != nil {
				return microerror.Mask(err)
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
				return microerror.Mask(err)
			}
			defer resp.Body.Close()

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return microerror.Mask(err)
			}

			var r E2EAppResponse
			err = json.Unmarshal(b, &r)
			if err != nil {
				return microerror.Mask(err)
			}

			if r.Name != e2eAppName {
				return microerror.Maskf(e2eAppError, "expctected name %#q got %#q", e2eAppName, r.Name)
			} else if r.Source != e2eAppSource {
				return microerror.Maskf(e2eAppError, "expctected name %#q got %#q", e2eAppSource, r.Source)
			}

			return nil
		}

		debugFailure := func(err error) {
			if err != nil {
				newLogger.Log("level", "warning", "message", "failed requesting e2e-app", "stack", fmt.Sprintf("%#v", err))
			}

			o := metav1.ListOptions{
				LabelSelector: "app=e2e-app",
			}
			l, err := config.Guest.K8sClient().CoreV1().Pods(ChartNamespace).List(o)
			if err != nil {
				t.Fatalf("expected %#v got %#v", nil, err)
			}

			for _, p := range l.Items {
				newLogger.Log("level", "debug", "message", fmt.Sprintf("e2e-app pod %#q has status %#v", p.GetName(), p.Status))
			}
		}

		go func() {
			newLogger.Log("level", "debug", "message", "requesting e2e-app continuously")
			defer newLogger.Log("level", "debug", "message", "requested e2e-app continuously")

			for {
				time.Sleep(100 * time.Millisecond)

				select {
				case <-done:
					return
				default:
					err := requestE2EApp()
					if err != nil {
						debugFailure(err)
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
		newLogger.Log("level", "debug", "message", "verified e2e-app availability 10 more seconds")
	}

	{
		newLogger.Log("level", "debug", "message", "scaling down tenant cluster worker")

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
		err = config.Guest.WaitForNodesReady(ctx, masterCount+workerCount-1)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		newLogger.Log("level", "debug", "message", "scaled down tenant cluster worker")
	}

	{
		newLogger.Log("level", "debug", "message", "verifying e2e-app availability 10 more seconds")
		time.Sleep(10 * time.Second)
		newLogger.Log("level", "debug", "message", "verified e2e-app availability 10 more seconds")
	}

	close(done)

	{
		newLogger.Log("level", "debug", "message", "validating test data")

		acceptable := float64(0)
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
	cluster, err := config.K8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.MasterCount(*cluster), nil
}

func numberOfWorkers(clusterID string) (int, error) {
	cluster, err := config.K8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(clusterID, metav1.GetOptions{})
	if err != nil {
		return 0, microerror.Mask(err)
	}

	return key.WorkerCount(*cluster), nil
}

func removeWorker(clusterID string) error {
	// TODO remove deprecated approach when v22 is gone.
	{
		patch := make([]framework.PatchSpec, 1)
		patch[0].Op = "remove"
		patch[0].Path = "/spec/aws/workers/1"

		patchBytes, err := json.Marshal(patch)
		if err != nil {
			return microerror.Mask(err)
		}
		_, err = config.K8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Patch(clusterID, types.JSONPatchType, patchBytes)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		customObject, err := config.K8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Get(clusterID, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}

		customObject.Spec.Cluster.Scaling.Max--
		customObject.Spec.Cluster.Scaling.Min--

		_, err = config.K8sClients.G8sClient().ProviderV1alpha1().AWSConfigs("default").Update(customObject)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

package loadtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/giantswarm/apiextensions/pkg/apis/core/v1alpha1"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/helm/pkg/helm"
)

type Config struct {
	Clients        *Clients
	GuestFramework *framework.Guest
	Logger         micrologger.Logger

	AuthToken    string
	ClusterID    string
	CommonDomain string
}

type LoadTest struct {
	clients        *Clients
	guestFramework *framework.Guest
	logger         micrologger.Logger

	authToken    string
	clusterID    string
	commonDomain string
}

func New(config Config) (*LoadTest, error) {
	if config.Clients.ControlPlaneHelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients.ControlPlaneHelmClient must not be empty", config)
	}
	if config.Clients.ControlPlaneK8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Clients.ControlPlaneK8sClient must not be empty", config)
	}
	if config.GuestFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.GuestFramework must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.AuthToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.AuthToken must not be empty", config)
	}
	if config.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", config)
	}
	if config.CommonDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CommonDomain must not be empty", config)
	}

	s := &LoadTest{
		clients:        config.Clients,
		guestFramework: config.GuestFramework,
		logger:         config.Logger,

		authToken:    config.AuthToken,
		clusterID:    config.ClusterID,
		commonDomain: config.CommonDomain,
	}

	return s, nil
}

func (l *LoadTest) Test(ctx context.Context) error {
	var err error

	var loadTestEndpoint string
	{
		loadTestEndpoint = fmt.Sprintf("loadtest-app.%s.%s", l.clusterID, l.commonDomain)

		l.logger.Log("level", "debug", "message", fmt.Sprintf("loadtest-app endpoint is %#q", loadTestEndpoint))
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "enabling HPA for Nginx Ingress Controller")

		err = l.enableIngressControllerHPA(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "enabled HPA for Nginx Ingress Controller")
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "installing loadtest app")

		err = l.installTestApp(ctx, loadTestEndpoint)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "installed loadtest app")
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "waiting for loadtest app to be ready")

		err = l.waitForLoadTestApp(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "loadtest app is ready")
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "installing loadtest job")

		err = l.installLoadTestJob(ctx, loadTestEndpoint)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "installed loadtest job")
	}

	var jsonResults []byte

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "waiting for loadtest job to complete")

		jsonResults, err = l.waitForLoadTestJob(ctx)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "loadtest job is complete")
	}

	{
		l.logger.LogCtx(ctx, "level", "debug", "message", "checking loadtest results")

		err = l.checkLoadTestResults(ctx, jsonResults)
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.LogCtx(ctx, "level", "debug", "message", "checked loadtest results")
	}

	return nil
}

// checkLoadTestResults parses the load test results JSON and determines if the
// test was successful or not.
func (l *LoadTest) checkLoadTestResults(ctx context.Context, jsonResults []byte) error {
	var err error

	l.logger.LogCtx(ctx, "level", "debug", "message", "checking loadtest results")

	l.logger.LogCtx(ctx, "level", "debug", "message", jsonResults)

	var results LoadTestResults

	err = json.Unmarshal(jsonResults, &results)
	if err != nil {
		return microerror.Mask(err)
	}

	apdexScore := results.Data.Attributes.BasicStatistics.Apdex75

	if apdexScore < ApdexPassThreshold {
		return microerror.Maskf(invalidExecutionError, "apdex score of %f is less than %f", apdexScore, ApdexPassThreshold)
	}

	l.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("load test passed: apdex score of %f is >= %f", apdexScore, ApdexPassThreshold))

	return nil
}

// enableIngressControllerHPA enables HPA via the user configmap and updates
// the chartconfig CR so chart-operator reconciles the config change.
func (l *LoadTest) enableIngressControllerHPA(ctx context.Context) error {
	var err error

	values := UserConfigMapValues{
		Data: UserConfigMapValuesData{
			AutoscalingEnabled: true,
		},
	}

	var data []byte

	data, err = yaml.Marshal(values)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = l.guestFramework.K8sClient().CoreV1().ConfigMaps(metav1.NamespaceSystem).Patch(UserConfigMapName, types.StrategicMergePatchType, data)
	if err != nil {
		return microerror.Mask(err)
	}

	var cr *v1alpha1.ChartConfig

	cr, err = l.guestFramework.G8sClient().CoreV1alpha1().ChartConfigs(CustomResourceNamespace).Get(CustomResourceName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	annotations := cr.Annotations
	annotations["test"] = "test"
	cr.SetAnnotations(annotations)

	_, err = l.guestFramework.G8sClient().CoreV1alpha1().ChartConfigs(CustomResourceNamespace).Update(cr)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// installLoadTestJob installs a chart that creates a job that uses the
// Stormforger CLI to trigger the load test.
func (l *LoadTest) installLoadTestJob(ctx context.Context, loadTestEndpoint string) error {
	var err error

	var jsonValues []byte
	{
		values := LoadTestValues{
			Auth: LoadTestValuesAuth{
				Token: l.authToken,
			},
			Test: LoadTestValuesTest{
				Endpoint: loadTestEndpoint,
				Name:     TestName,
			},
		}

		jsonValues, err = json.Marshal(values)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = l.installChart(ctx, l.clients.ControlPlaneHelmClient, JobChartName, jsonValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

// installLoadTestApp installs a chart that deploys the Stormforger test app
// in the tenant cluster as the test workload for the load test.
func (l *LoadTest) installTestApp(ctx context.Context, loadTestEndpoint string) error {
	var err error

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

	var tenantHelmClient helmclient.Interface

	{
		c := helmclient.Config{
			Logger:    l.logger,
			K8sClient: l.guestFramework.K8sClient(),

			RestConfig: l.guestFramework.RestConfig(),
		}

		tenantHelmClient, err = helmclient.New(c)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = tenantHelmClient.EnsureTillerInstalled(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	{
		err = l.installChart(ctx, tenantHelmClient, AppChartName, jsonValues)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

// waitForLoadTestApp waits for all pods of the test app to be ready.
func (l *LoadTest) waitForLoadTestApp(ctx context.Context) error {
	l.logger.Log("level", "debug", "message", "waiting for loadtest-app deployment to be ready")

	o := func() error {
		lo := metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=loadtest-app",
		}
		l, err := l.guestFramework.K8sClient().AppsV1().Deployments(metav1.NamespaceDefault).List(lo)
		if err != nil {
			return microerror.Mask(err)
		}
		if len(l.Items) != 1 {
			return microerror.Maskf(waitError, "want %d deployments found %d", 1, len(l.Items))
		}

		deploy := l.Items[0]
		if *deploy.Spec.Replicas == deploy.Status.ReadyReplicas {
			return microerror.Maskf(waitError, "want %d ready pods found %d", deploy.Spec.Replicas, deploy.Status.ReadyReplicas)
		}

		return nil
	}

	b := backoff.NewConstant(2*time.Minute, 15*time.Second)
	n := func(err error, delay time.Duration) {
		l.logger.Log("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return microerror.Mask(err)
	}

	l.logger.Log("level", "debug", "message", "waited for loadtest-app deployment to be ready")

	return nil
}

// waitForLoadTestJob waits for the job running the Stormforger CLI to
// complete and then gets the pod logs which contains the results JSON. The CLI
// is configured to wait for the load test to complete.
func (l *LoadTest) waitForLoadTestJob(ctx context.Context) ([]byte, error) {
	var podCount = 1
	var podName = ""

	l.logger.Log("level", "debug", "message", "waiting for stormforger-cli job")

	o := func() error {
		lo := metav1.ListOptions{
			FieldSelector: "status.phase=Succeeded",
			LabelSelector: "app.kubernetes.io/name=stormforger-cli",
		}
		l, err := l.clients.ControlPlaneK8sClient.CoreV1().Pods(metav1.NamespaceDefault).List(lo)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(l.Items) == podCount {
			podName = l.Items[0].Name

			return nil
		}

		return microerror.Maskf(waitError, "want %d Succeeded pods found %d", podCount, len(l.Items))
	}

	b := backoff.NewConstant(20*time.Minute, 30*time.Second)
	n := func(err error, delay time.Duration) {
		l.logger.Log("level", "debug", "message", err.Error())
	}

	err := backoff.RetryNotify(o, b, n)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	l.logger.Log("level", "debug", "message", "waited for stormforger-cli job")

	l.logger.Log("level", "debug", "message", "getting results from pod logs")

	req := l.clients.ControlPlaneK8sClient.CoreV1().Pods(metav1.NamespaceDefault).GetLogs(podName, &corev1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		return nil, err
	}

	defer readCloser.Close()

	buf := new(bytes.Buffer)

	_, err = io.Copy(buf, readCloser)
	if err != nil {
		return nil, err
	}

	l.logger.Log("level", "debug", "message", "got results from pod logs")

	return buf.Bytes(), nil
}

// installChart is a helper method for installing helm charts.
func (l *LoadTest) installChart(ctx context.Context, helmClient helmclient.Interface, chartName string, jsonValues []byte) error {
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

	{
		l.logger.Log("level", "debug", "message", fmt.Sprintf("installing %#q", chartName))

		tarballPath, err := apprClient.PullChartTarball(ctx, chartName, ChartChannel)
		if err != nil {
			return microerror.Mask(err)
		}

		err = helmClient.InstallReleaseFromTarball(ctx, tarballPath, ChartNamespace, helm.ValueOverrides(jsonValues))
		if err != nil {
			return microerror.Mask(err)
		}

		l.logger.Log("level", "debug", "message", fmt.Sprintf("installed %#q", chartName))
	}

	return nil
}

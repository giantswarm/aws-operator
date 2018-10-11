package release

import (
	"context"
	"fmt"
	"time"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/apprclient"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/pkg/framework/filelogger"
	"github.com/giantswarm/helmclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/helm"
)

const (
	defaultNamespace = "default"
)

type Config struct {
	ApprClient *apprclient.Client
	ExtClient  apiextensionsclient.Interface
	FileLogger *filelogger.FileLogger
	G8sClient  versioned.Interface
	HelmClient *helmclient.Client
	K8sClient  kubernetes.Interface
	Logger     micrologger.Logger

	Namespace string
}

type Release struct {
	apprClient *apprclient.Client
	fileLogger *filelogger.FileLogger
	helmClient *helmclient.Client
	k8sClient  kubernetes.Interface
	logger     micrologger.Logger

	condition *conditionSet
	namespace string
}

func New(config Config) (*Release, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.ApprClient == nil {
		config.Logger.Log("level", "debug", "message", fmt.Sprintf("%T.ApprClient is empty", config))

		config.Logger.Log("level", "debug", "message", fmt.Sprintf("using default for %T.ApprClient", config))

		c := apprclient.Config{
			Fs:     afero.NewOsFs(),
			Logger: config.Logger,

			Address:      "https://quay.io",
			Organization: "giantswarm",
		}

		a, err := apprclient.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		config.ApprClient = a
	}
	if config.FileLogger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.FileLogger must not be empty", config)
	}
	if config.ExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ExtClient must not be empty", config)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.G8sClient must not be empty", config)
	}
	if config.HelmClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.HelmClient must not be empty", config)
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Namespace == "" {
		config.Namespace = defaultNamespace
	}

	var err error

	var condition *conditionSet
	{

		c := conditionSetConfig{
			ExtClient: config.ExtClient,
			Logger:    config.Logger,
		}

		condition, err = newConditionSet(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

	}

	r := &Release{
		apprClient: config.ApprClient,
		fileLogger: config.FileLogger,
		helmClient: config.HelmClient,
		k8sClient:  config.K8sClient,
		logger:     config.Logger,

		condition: condition,
		namespace: config.Namespace,
	}

	return r, nil
}

func (r *Release) Delete(ctx context.Context, name string) error {
	err := r.helmClient.DeleteRelease(name, helm.DeletePurge(true))
	if helmclient.IsReleaseNotFound(err) {
		return microerror.Maskf(releaseNotFoundError, name)
	} else if helmclient.IsTillerNotFound(err) {
		return microerror.Mask(tillerNotFoundError)
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Release) EnsureDeleted(ctx context.Context, name string) error {
	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring deletion of release %#q", name))

	err := r.helmClient.DeleteRelease(name, helm.DeletePurge(true))
	if helmclient.IsReleaseNotFound(err) {
		r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("release %#q does not exist", name))
	} else if helmclient.IsTillerNotFound(err) {
		r.logger.LogCtx(ctx, "level", "warning", "message", "tiller is not found/installed")
	} else if err != nil {
		return microerror.Mask(err)
	} else {
		r.logger.LogCtx(ctx, "level", "info", "message", fmt.Sprintf("deleted release %#q", name))
	}

	r.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured deletion of release %#q", name))

	return nil
}

func (r *Release) Install(ctx context.Context, name string, version Version, values string, conditions ...func() error) error {
	var err error

	chartname := fmt.Sprintf("%s-chart", name)

	var tarball string
	if version.isChannel {
		tarball, err = r.apprClient.PullChartTarball(chartname, version.String())
		if err != nil {
			return microerror.Mask(err)
		}
	} else {
		tarball, err = r.apprClient.PullChartTarballFromRelease(chartname, version.String())
		if err != nil {
			return microerror.Mask(err)
		}
	}

	err = r.helmClient.InstallFromTarball(tarball, r.namespace, helm.ReleaseName(name), helm.ValueOverrides([]byte(values)), helm.InstallWait(true))
	if err != nil {
		return microerror.Mask(err)
	}

	for _, c := range conditions {
		err = backoff.Retry(c, backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval))
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (r *Release) InstallOperator(ctx context.Context, name string, version Version, values string, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	err := r.Install(ctx, name, version, values, r.condition.CRD(ctx, crd))
	if err != nil {
		return microerror.Mask(err)
	}
	// TODO introduced: https://github.com/giantswarm/e2e-harness/pull/121
	// This fallback from r.namespace was introduced because not all our
	// operators accept and apply configured namespaces.
	//
	// Tracking issue: https://github.com/giantswarm/giantswarm/issues/4123
	//
	// Final version of the code:
	//
	//	podName, err := r.podName(r.namespace, fmt.Sprintf("app=%s", name))
	//	if err != nil {
	//		return microerror.Mask(err)
	//	}
	//	err = r.filelogger.StartLoggingPod(r.namespace, podName)
	//	if err != nil {
	//		return microerror.Mask(err)
	//	}
	//
	podNamespace := r.namespace

	podName, err := r.podName(podNamespace, fmt.Sprintf("app=%s", name))
	if IsNotFound(err) {
		podNamespace = "giantswarm"
		podName, err = r.podName(podNamespace, fmt.Sprintf("app=%s", name))
		if err != nil {
			return microerror.Mask(err)
		}
	} else if err != nil {
		return microerror.Mask(err)
	}

	err = r.fileLogger.StartLoggingPod(podNamespace, podName)
	if err != nil {
		return microerror.Mask(err)
	}
	// TODO end

	return nil
}

func (r *Release) Update(ctx context.Context, name, values, channel string, conditions ...func() error) error {
	chartname := fmt.Sprintf("%s-chart", name)

	tarballPath, err := r.apprClient.PullChartTarball(chartname, channel)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.helmClient.UpdateReleaseFromTarball(name, tarballPath, helm.UpdateValueOverrides([]byte(values)))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Release) WaitForStatus(ctx context.Context, release string, status string) error {
	operation := func() error {
		rc, err := r.helmClient.GetReleaseContent(release)
		if helmclient.IsReleaseNotFound(err) && status == "DELETED" {
			// Error is expected because we purge releases when deleting.
			return nil
		} else if err != nil {
			return microerror.Mask(err)
		}
		if rc.Status != status {
			return microerror.Maskf(releaseStatusNotMatchingError, "waiting for '%s', current '%s'", status, rc.Status)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("failed to get release status '%s': retrying in %s", status, t), "stack", fmt.Sprintf("%v", err))
	}

	b := backoff.NewExponential(backoff.MediumMaxWait, backoff.LongMaxInterval)
	err := backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (r *Release) WaitForVersion(ctx context.Context, release string, version string) error {
	operation := func() error {
		rh, err := r.helmClient.GetReleaseHistory(release)
		if err != nil {
			return microerror.Mask(err)
		}
		if rh.Version != version {
			return microerror.Maskf(releaseVersionNotMatchingError, "waiting for '%s', current '%s'", version, rh.Version)
		}
		return nil
	}

	notify := func(err error, t time.Duration) {
		r.logger.Log("level", "debug", "message", fmt.Sprintf("failed to get release version '%s': retrying in %s", version, t), "stack", fmt.Sprintf("%v", err))
	}

	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.LongMaxInterval)
	err := backoff.RetryNotify(operation, b, notify)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (r *Release) podName(namespace, labelSelector string) (string, error) {
	pods, err := r.k8sClient.CoreV1().
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

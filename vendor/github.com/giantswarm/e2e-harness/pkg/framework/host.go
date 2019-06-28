package framework

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/backoff"
	"github.com/giantswarm/e2e-harness/internal/filelogger"
	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	aggregationclient "k8s.io/kube-aggregator/pkg/client/clientset_generated/clientset"
)

const (
	defaultNamespace = "default"
)

type HostConfig struct {
	Backoff backoff.Interface
	Logger  micrologger.Logger

	ClusterID       string
	TargetNamespace string
}

type Host struct {
	backoff    backoff.Interface
	logger     micrologger.Logger
	filelogger *filelogger.FileLogger

	extClient            *apiextensionsclient.Clientset
	g8sClient            *versioned.Clientset
	k8sClient            *kubernetes.Clientset
	k8sAggregationClient *aggregationclient.Clientset
	restConfig           *rest.Config

	clusterID       string
	targetNamespace string
}

func NewHost(c HostConfig) (*Host, error) {
	if c.Backoff == nil {
		c.Backoff = backoff.NewExponential(backoff.ShortMaxWait, backoff.LongMaxInterval)
	}
	if c.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	if c.ClusterID == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterID must not be empty", c)
	}
	if c.TargetNamespace == "" {
		c.TargetNamespace = defaultNamespace
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	extClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	g8sClient, err := versioned.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	k8sAggregationClient, err := aggregationclient.NewForConfig(restConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	var fileLogger *filelogger.FileLogger
	{
		fc := filelogger.Config{
			K8sClient: k8sClient,
			Logger:    c.Logger,
		}
		fileLogger, err = filelogger.New(fc)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	h := &Host{
		backoff:    c.Backoff,
		logger:     c.Logger,
		filelogger: fileLogger,

		extClient:            extClient,
		g8sClient:            g8sClient,
		k8sClient:            k8sClient,
		k8sAggregationClient: k8sAggregationClient,
		restConfig:           restConfig,

		clusterID:       c.ClusterID,
		targetNamespace: c.TargetNamespace,
	}

	return h, nil
}

func (h *Host) ApplyAWSConfigPatch(patch []PatchSpec, clusterName string) error {
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = h.g8sClient.
		ProviderV1alpha1().
		AWSConfigs(h.targetNamespace).
		Patch(clusterName, types.JSONPatchType, patchBytes)

	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (h *Host) AWSCluster(name string) (*v1alpha1.AWSConfig, error) {
	cluster, err := h.g8sClient.ProviderV1alpha1().
		AWSConfigs(h.targetNamespace).
		Get(name, metav1.GetOptions{})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cluster, nil
}

func (h *Host) DeleteGuestCluster(ctx context.Context, provider string) error {
	{
		h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("triggering deletion of CR for guest cluster %#q", h.clusterID))

		o := func() error {
			var err error

			switch provider {
			case "aws":
				err = h.g8sClient.ProviderV1alpha1().AWSConfigs(h.targetNamespace).Delete(h.clusterID, &metav1.DeleteOptions{})
			case "azure":
				err = h.g8sClient.ProviderV1alpha1().AzureConfigs(h.targetNamespace).Delete(h.clusterID, &metav1.DeleteOptions{})
			case "kvm":
				err = h.g8sClient.ProviderV1alpha1().KVMConfigs(h.targetNamespace).Delete(h.clusterID, &metav1.DeleteOptions{})
			default:
				return microerror.Maskf(unknownProviderError, "%#q not recognized", provider)
			}

			if err != nil {
				return microerror.Mask(err)
			}
			return nil
		}

		n := backoff.NewNotifier(h.logger, context.Background())
		err := backoff.RetryNotify(o, h.backoff, n)
		if apierrors.IsNotFound(err) {
			h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not trigger deletion of CR for guest cluster %#q", h.clusterID))
			h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("CR for guest cluster %#q does not exist", h.clusterID))
		} else if err != nil {
			h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not trigger deletion of CR for guest cluster %#q", h.clusterID))
			return microerror.Mask(err)
		}

		h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("triggered deletion of CR for guest cluster %#q", h.clusterID))
	}

	{
		h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring deletion of CR for guest cluster %#q", h.clusterID))

		o := func() error {
			var err error

			switch provider {
			case "aws":
				_, err = h.g8sClient.ProviderV1alpha1().AWSConfigs(h.targetNamespace).Get(h.clusterID, metav1.GetOptions{})
			case "azure":
				_, err = h.g8sClient.ProviderV1alpha1().AzureConfigs(h.targetNamespace).Get(h.clusterID, metav1.GetOptions{})
			case "kvm":
				_, err = h.g8sClient.ProviderV1alpha1().KVMConfigs(h.targetNamespace).Get(h.clusterID, metav1.GetOptions{})
			default:
				return microerror.Maskf(unknownProviderError, "%#q not recognized", provider)
			}

			if apierrors.IsNotFound(err) {
				return nil
			} else if err != nil {
				return microerror.Mask(err)
			} else {
				return microerror.Maskf(clusterDeletionError, "guest cluster %#q CR still exists", h.clusterID)
			}
		}

		b := backoff.NewExponential(backoff.LongMaxWait, backoff.LongMaxInterval)
		n := backoff.NewNotifier(h.logger, context.Background())
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("did not ensure deletion of CR for guest cluster %#q", h.clusterID))
			return microerror.Mask(err)
		}

		h.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured deletion of CR for guest cluster %#q", h.clusterID))
	}

	return nil
}

func (h *Host) ExtClient() apiextensionsclient.Interface {
	return h.extClient
}

// G8sClient returns the host cluster framework's Giant Swarm client.
func (h *Host) G8sClient() versioned.Interface {
	return h.g8sClient
}

// K8sClient returns the host cluster framework's Kubernetes client.
func (h *Host) K8sClient() kubernetes.Interface {
	return h.k8sClient
}

// K8sAggregationClient returns the host cluster framework's Kubernetes aggregation client.
func (h *Host) K8sAggregationClient() *aggregationclient.Clientset {
	return h.k8sAggregationClient
}

// RestConfig returns the host cluster framework's rest config.
func (h *Host) RestConfig() *rest.Config {
	return h.restConfig
}

func (h *Host) TargetNamespace() string {
	return h.targetNamespace
}

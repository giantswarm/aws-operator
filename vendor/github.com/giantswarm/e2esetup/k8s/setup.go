package k8s

import (
	"context"
	"fmt"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type SetupConfig struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

type Setup struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func NewSetup(config SetupConfig) (*Setup, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	s := &Setup{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}

	return s, nil
}

func (s *Setup) EnsureNamespaceCreated(ctx context.Context, namespace string) error {
	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring Kubernetes Namespace %#q", namespace))

	o := func() error {
		{
			n := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			_, err := s.k8sClient.CoreV1().Namespaces().Create(n)
			if errors.IsAlreadyExists(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		{
			n, err := s.k8sClient.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			if n.Status.Phase != v1.NamespaceActive {
				return microerror.Maskf(unexpectedStatusPhaseError, string(n.Status.Phase))
			}
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

	err := backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured Kubernetes Namespace %#q", namespace))

	return nil
}

func (s *Setup) EnsureNamespaceDeleted(ctx context.Context, namespace string) error {
	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensuring deletion of Kubernetes Namespace %#q", namespace))

	o := func() error {
		{
			err := s.k8sClient.CoreV1().Namespaces().Delete(namespace, &metav1.DeleteOptions{})
			if errors.IsNotFound(err) {
				// fall through
			} else if err != nil {
				return microerror.Mask(err)
			}
		}

		return nil
	}
	b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)

	err := backoff.Retry(o, b)
	if err != nil {
		return microerror.Mask(err)
	}

	s.logger.LogCtx(ctx, "level", "debug", "message", fmt.Sprintf("ensured deletion of Kubernetes Namespace %#q", namespace))

	return nil
}

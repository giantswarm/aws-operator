package release

import (
	"context"

	"github.com/giantswarm/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type conditionFn func() error

type conditionSetConfig struct {
	ExtClient apiextensionsclient.Interface
	Logger    micrologger.Logger
}

type conditionSet struct {
	extClient apiextensionsclient.Interface
	logger    micrologger.Logger
}

func newConditionSet(config conditionSetConfig) (*conditionSet, error) {
	if config.ExtClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.ExtClient must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	c := &conditionSet{
		extClient: config.ExtClient,
		logger:    config.Logger,
	}

	return c, nil
}

func (c *conditionSet) CRD(ctx context.Context, crd *apiextensionsv1beta1.CustomResourceDefinition) conditionFn {
	return func() error {
		o := func() error {
			_, err := c.extClient.ApiextensionsV1beta1().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return microerror.Mask(err)
			} else if err != nil {
				return backoff.Permanent(microerror.Mask(err))
			}
			return nil
		}
		b := backoff.NewExponential(backoff.ShortMaxWait, backoff.ShortMaxInterval)
		n := backoff.NewNotifier(c.logger, ctx)
		err := backoff.RetryNotify(o, b, n)
		if err != nil {
			return microerror.Mask(err)
		}

		return nil
	}
}

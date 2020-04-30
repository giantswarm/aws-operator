package hamaster

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type HAMaster struct {
	k8sClient k8sclient.Interface
}

func New(config Config) (*HAMaster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	h := &HAMaster{
		k8sClient: config.K8sClient,
	}

	return h, nil
}

func (h *HAMaster) Enabled(ctx context.Context, cluster string) (bool, error) {
	var list infrastructurev1alpha2.G8sControlPlaneList

	err := h.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(v1.NamespaceDefault),
		client.MatchingLabels{label.Cluster: cluster},
	)
	if err != nil {
		return false, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return false, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return false, microerror.Mask(tooManyCRsError)
	}

	if key.G8sControlPlaneReplicas(list.Items[0]) == 1 {
		return false, nil
	}

	return true, nil
}

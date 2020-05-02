package hamaster

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type HAMaster struct {
	k8sClient k8sclient.Interface

	azs []string
	ids []int
	ptr int
}

func New(config Config) (*HAMaster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	h := &HAMaster{
		k8sClient: config.K8sClient,

		azs: []string{},
		ids: []int{},
	}

	return h, nil
}

func (h *HAMaster) AZ() string {
	return h.azs[h.ptr]
}

func (h *HAMaster) ID() int {
	return h.ids[h.ptr]
}

func (h *HAMaster) Init(ctx context.Context, obj interface{}) error {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return microerror.Mask(err)
	}

	// We need the G8sControlPlane CR because it holds the replica count. This
	// tells us how many masters the current setup defines and ultimately dictates
	// the Master IDs. The system's implementation requires there only to be 1 or
	// 3 masters.
	var g8s infrastructurev1alpha2.G8sControlPlane
	{
		var list infrastructurev1alpha2.G8sControlPlaneList

		err := h.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(list.Items) == 0 {
			return microerror.Mask(notFoundError)
		}
		if len(list.Items) > 1 {
			return microerror.Mask(tooManyCRsError)
		}

		g8s = list.Items[0]
	}

	// We need the AWSControlPlane CR because it holds the availability zones. The
	// state machine allows to cycle through them in a deterministic way. The
	// system's implementation requires there only to be 1, 2 or 3 availability
	// zones.
	var aws infrastructurev1alpha2.AWSControlPlane
	{
		var list infrastructurev1alpha2.AWSControlPlaneList

		err := h.k8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace(cr.GetNamespace()),
			client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
		)
		if err != nil {
			return microerror.Mask(err)
		}

		if len(list.Items) == 0 {
			return microerror.Mask(notFoundError)
		}
		if len(list.Items) > 1 {
			return microerror.Mask(tooManyCRsError)
		}

		aws = list.Items[0]
	}

	// We need a deterministic list of availability zones which we can cycle
	// through for the required amount of masters. Eventually it happens that
	// there is only 1 availability zone in a HA Masters setup. Therefore the
	// internal state holds the given availability zones repeatedly 3 times so
	// that it can always guarantee an availability zone for each master in any of
	// the allowed permutations. Note that given 3 availability zones in a HA
	// Masters setup will never make use of the repeated availability zones, which
	// is just a side effect of the current implementation.
	{
		h.azs = append(h.azs, aws.Spec.AvailabilityZones...)
		h.azs = append(h.azs, aws.Spec.AvailabilityZones...)
		h.azs = append(h.azs, aws.Spec.AvailabilityZones...)
	}

	// The master IDs are only allowed to be either 0, 1, 2 or 3, so we hard code
	// the list of master IDs depending on the given replicas, which must be
	// either 1 or 3.
	{
		if g8s.Spec.Replicas == 1 {
			h.ids = []int{0}
		}
		if g8s.Spec.Replicas == 3 {
			h.ids = []int{1, 2, 3}
		}
	}

	return nil
}

func (h *HAMaster) Next() {
	if h.ptr == len(h.ids) {
		h.ptr = 0
	} else {
		h.ptr++
	}
}

func (h *HAMaster) Reconciled() bool {
	if h.ptr == len(h.ids) {
		return true
	} else {
		return false
	}
}

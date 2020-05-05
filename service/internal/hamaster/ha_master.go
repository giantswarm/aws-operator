package hamaster

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/cachekeycontext"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type Config struct {
	K8sClient k8sclient.Interface
}

type HAMaster struct {
	k8sClient k8sclient.Interface

	awsCache map[string]infrastructurev1alpha2.AWSControlPlane
	g8sCache map[string]infrastructurev1alpha2.G8sControlPlane
}

func New(config Config) (*HAMaster, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	h := &HAMaster{
		k8sClient: config.K8sClient,

		awsCache: map[string]infrastructurev1alpha2.AWSControlPlane{},
		g8sCache: map[string]infrastructurev1alpha2.G8sControlPlane{},
	}

	return h, nil
}

func (h *HAMaster) Mapping(ctx context.Context, obj interface{}) ([]Mapping, error) {
	var err error
	var ok bool

	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var cacheKey string
	{
		ck, ok := cachekeycontext.FromContext(ctx)
		if ok {
			cacheKey = fmt.Sprintf("%s/%s", ck, key.ClusterID(cr))
		}
	}

	// We need the G8sControlPlane CR because it holds the replica count. This
	// tells us how many masters the current setup defines and ultimately dictates
	// the Master IDs. The system's implementation requires there only to be 1 or
	// 3 masters.
	var g8s infrastructurev1alpha2.G8sControlPlane
	if cacheKey == "" {
		g8s, err = h.getG8s(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		g8s, ok = h.g8sCache[cacheKey]
		if !ok {
			g8s, err = h.getG8s(ctx, cr)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if len(h.g8sCache) == 1 {
				h.g8sCache = map[string]infrastructurev1alpha2.G8sControlPlane{}
			}

			h.g8sCache[cacheKey] = g8s
		}
	}

	// We need the AWSControlPlane CR because it holds the availability zones. The
	// system's implementation requires there only to be 1, 2 or 3 availability
	// zones.
	var aws infrastructurev1alpha2.AWSControlPlane
	if cacheKey == "" {
		aws, err = h.getAWS(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	} else {
		aws, ok = h.awsCache[cacheKey]
		if !ok {
			aws, err = h.getAWS(ctx, cr)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			if len(h.awsCache) == 1 {
				h.awsCache = map[string]infrastructurev1alpha2.AWSControlPlane{}
			}

			h.awsCache[cacheKey] = aws
		}
	}

	// We need a deterministic list of availability zones which we can loop over
	// through for the required amount of masters. Eventually it happens that
	// there is only 1 availability zone in a HA Masters setup. Therefore the
	// computed mapping holds the given availability zones repeatedly 3 times so
	// that there is always a guaranteed availability zone for each master in any
	// of the allowed permutations. Note that given 3 availability zones in a HA
	// Masters setup we will never make use of the repeated availability zones,
	// which is just a side effect of the current implementation.
	var azs []string
	{
		azs = append(azs, aws.Spec.AvailabilityZones...)
		azs = append(azs, aws.Spec.AvailabilityZones...)
		azs = append(azs, aws.Spec.AvailabilityZones...)
	}

	// The master IDs are only allowed to be either 0, 1, 2 or 3, so we hard code
	// the list of master IDs depending on the given replicas, which must be
	// either 1 or 3.
	var ids []int
	{
		if key.G8sControlPlaneReplicas(g8s) == 1 {
			ids = append(ids, 0)
		}
		if key.G8sControlPlaneReplicas(g8s) == 3 {
			ids = append(ids, 1)
			ids = append(ids, 2)
			ids = append(ids, 3)
		}
	}

	var mappings []Mapping
	for i, _ := range ids {
		m := Mapping{
			AZ: azs[i],
			ID: ids[i],
		}

		mappings = append(mappings, m)
	}

	return mappings, nil
}

func (h *HAMaster) getAWS(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.AWSControlPlane, error) {
	var list infrastructurev1alpha2.AWSControlPlaneList

	err := h.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(cr.GetNamespace()),
		client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
	)
	if err != nil {
		return infrastructurev1alpha2.AWSControlPlane{}, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return infrastructurev1alpha2.AWSControlPlane{}, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return infrastructurev1alpha2.AWSControlPlane{}, microerror.Mask(tooManyCRsError)
	}

	return list.Items[0], nil
}

func (h *HAMaster) getG8s(ctx context.Context, cr metav1.Object) (infrastructurev1alpha2.G8sControlPlane, error) {
	var list infrastructurev1alpha2.G8sControlPlaneList

	err := h.k8sClient.CtrlClient().List(
		ctx,
		&list,
		client.InNamespace(cr.GetNamespace()),
		client.MatchingLabels{label.Cluster: key.ClusterID(cr)},
	)
	if err != nil {
		return infrastructurev1alpha2.G8sControlPlane{}, microerror.Mask(err)
	}

	if len(list.Items) == 0 {
		return infrastructurev1alpha2.G8sControlPlane{}, microerror.Mask(notFoundError)
	}
	if len(list.Items) > 1 {
		return infrastructurev1alpha2.G8sControlPlane{}, microerror.Mask(tooManyCRsError)
	}

	return list.Items[0], nil
}

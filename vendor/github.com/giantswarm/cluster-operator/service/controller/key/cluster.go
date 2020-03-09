package key

import (
	"fmt"

	"github.com/giantswarm/microerror"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

func APIEndpoint(cluster apiv1alpha2.Cluster, base string) string {
	return fmt.Sprintf("api.%s.k8s.%s", ClusterID(&cluster), base)
}

func TenantEndpoint(cluster apiv1alpha2.Cluster, base string) string {
	return fmt.Sprintf("%s.k8s.%s", ClusterID(&cluster), base)
}

func ToCluster(v interface{}) (apiv1alpha2.Cluster, error) {
	if v == nil {
		return apiv1alpha2.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha2.Cluster{}, v)
	}

	p, ok := v.(*apiv1alpha2.Cluster)
	if !ok {
		return apiv1alpha2.Cluster{}, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1alpha2.Cluster{}, v)
	}

	c := p.DeepCopy()

	return *c, nil
}

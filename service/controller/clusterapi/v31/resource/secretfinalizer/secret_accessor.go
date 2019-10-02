package secretfinalizer

import (
	"context"
	"fmt"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/key"
)

type secretAccessor struct {
	Name      string
	Namespace string
}

func newSecretAccessors(ctx context.Context, cr v1alpha1.Cluster) []secretAccessor {
	return []secretAccessor{
		// The secret accessors below are associated to the tenant's API
		// certificate.
		{
			Name:      fmt.Sprintf("%s-api", key.ClusterID(&cr)),
			Namespace: "default",
		},
		// The secret accessors below are associated to the tenant's BYOC
		// credential.
		{
			Name:      fmt.Sprintf("credential-%s", key.ClusterID(&cr)),
			Namespace: "giantswarm",
		},
	}
}

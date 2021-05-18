package secretfinalizer

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type secretAccessor struct {
	Name      string
	Namespace string
}

func newSecretAccessors(ctx context.Context, cr infrastructurev1alpha2.AWSCluster) []secretAccessor {
	return []secretAccessor{
		// The secret accessors below are associated to the tenant's API
		// certificate.
		{
			Name:      fmt.Sprintf("%s-api", key.ClusterID(&cr)),
			Namespace: cr.GetNamespace(),
		},
		// The secret accessors below are associated to the tenant's BYOC
		// credential.
		{
			Name:      fmt.Sprintf("credential-%s", key.ClusterID(&cr)),
			Namespace: "giantswarm",
		},
	}
}

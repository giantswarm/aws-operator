package secretfinalizer

import (
	"context"
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

type secretAccessor struct {
	Name      string
	Namespace string
}

func newSecretAccessors(ctx context.Context, cr v1alpha1.AWSConfig) []secretAccessor {
	return []secretAccessor{
		// The secret accessors below are associated to the tenant's API
		// certificate.
		{
			Name:      fmt.Sprintf("%s-api", key.ClusterID(cr)),
			Namespace: "default",
		},
		// The secret accessors below are associated to the tenant's BYOC
		// credential.
		{
			Name:      fmt.Sprintf("credential-%s", key.ClusterID(cr)),
			Namespace: "giantswarm",
		},
	}
}

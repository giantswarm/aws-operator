// +build k8srequired

package sameclusterid

import (
	"context"
	"testing"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

// Test_Recreate_Cluster makes sure we can create the cluster for the same CR
// after previous one is deleted. This improves coverage of resources
// idempotentency.
func Test_Recreate_Cluster(t *testing.T) {
	ctx := context.Background()

	{
		wait := true
		err := setup.EnsureTenantClusterDeleted(ctx, env.ClusterID(), config, wait)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

	{
		wait := true
		err := setup.EnsureTenantClusterCreated(ctx, env.ClusterID(), config, wait)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}
}

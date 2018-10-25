// +build k8srequired

package sameclusterid

import (
	"context"
	"testing"

	"github.com/giantswarm/aws-operator/integration/setup"
)

// Test_Recreate_Cluster makes sure we can create the cluster for the same CR
// after previous one is deleted. This improves coverage of resources
// idempotentency.
func Test_Recreate_Cluster(t *testing.T) {
	var err error
	ctx := context.Background()

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "deleting tenant cluster")

		err = setup.EnsureTenantClusterDeleted(ctx, config)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "deleted tenant cluster")
	}

	{
		config.Logger.LogCtx(ctx, "level", "debug", "message", "creating tenant cluster")

		err = setup.EnsureTenantClusterDeleted(ctx, config)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		config.Logger.LogCtx(ctx, "level", "debug", "message", "created tenant cluster")
	}
}

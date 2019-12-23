// +build k8srequired

package sameclusterid

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/giantswarm/aws-operator/integration/env"
	"github.com/giantswarm/aws-operator/integration/setup"
)

// Test_Recreate_Cluster makes sure we can create the cluster for the same CR
// after previous one is deleted. This improves coverage of resources
// idempotentency.
func Test_Recreate_Cluster(t *testing.T) {
	ctx := context.Background()

	var networkInterface *ec2.InstanceNetworkInterface
	if setup.BastionEnabled {
		var err error
		networkInterface, err = setup.RemoveBastionTenantAssociation(ctx, config, env.ClusterID())
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}

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

	if setup.BastionEnabled {
		err := setup.RestoreBastionTenantAssociation(ctx, config, networkInterface)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
	}
}

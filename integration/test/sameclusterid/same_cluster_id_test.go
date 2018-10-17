// +build k8srequired

package clusterstate

import (
	"context"
	"testing"
)

// Test_Same_Cluster_ID makes sure we can create the cluster with the same ID
// after previous one is deleted. This improves coverage of resources
// idempotentency.
func Test_Same_Cluster_ID(t *testing.T) {
	var err error
	ctx := context.Background()

	err = setup.DeleteTenantCluster(ctx, config)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}

	err = setup.CreateTenantCluster(ctx, config)
	if err != nil {
		t.Fatalf("expected %#v got %#v", nil, err)
	}
}

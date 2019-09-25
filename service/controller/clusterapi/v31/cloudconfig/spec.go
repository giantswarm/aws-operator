package cloudconfig

import (
	"context"

	"github.com/giantswarm/certs"
	"github.com/giantswarm/randomkeys"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

type Interface interface {
	Render(ctx context.Context, cr v1alpha1.Cluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, labels string) ([]byte, error)
}

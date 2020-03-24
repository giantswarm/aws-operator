package cloudconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/randomkeys"
)

type Interface interface {
	Render(ctx context.Context, g8client versioned.Interface, obj interface{}, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, labels string) ([]byte, error)
}

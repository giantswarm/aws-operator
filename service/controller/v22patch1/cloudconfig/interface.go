package cloudconfig

import (
	"context"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/randomkeys"
)

type Interface interface {
	NewMasterTemplate(ctx context.Context, customObject v1alpha1.AWSConfig, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster) (string, error)
	NewWorkerTemplate(ctx context.Context, customObject v1alpha1.AWSConfig, clusterCerts certs.Cluster) (string, error)
}

package cloudconfig

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/randomkeys"
)

type Interface interface {
	NewMasterTemplate(customObject v1alpha1.AWSConfig, certs legacy.CompactTLSAssets, clusterKeys randomkeys.Cluster, kmsKeyARN string) (string, error)
	NewWorkerTemplate(customObject v1alpha1.AWSConfig, certs legacy.CompactTLSAssets) (string, error)
}

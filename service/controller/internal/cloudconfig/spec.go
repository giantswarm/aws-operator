package cloudconfig

import (
	"context"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/randomkeys"
)

type Interface interface {
	RenderTCNP(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, images k8scloudconfig.Images, labels string) ([]byte, error)
	RenderTCCPN(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, images k8scloudconfig.Images, masterID int, labels string) ([]byte, error)
}

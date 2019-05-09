package kms

import (
	"fmt"

	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

func keyAlias(cr v1alpha1.Cluster) string {
	return fmt.Sprintf("alias/%s", key.ClusterID(cr))
}

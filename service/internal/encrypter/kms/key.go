package kms

import (
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func keyAlias(cr infrastructurev1alpha2.AWSCluster) string {
	return fmt.Sprintf("alias/%s", key.ClusterID(&cr))
}

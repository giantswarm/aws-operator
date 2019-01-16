package kms

import (
	"fmt"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/v21/key"
)

func keyAlias(customObject v1alpha1.AWSConfig) string {
	clusterID := key.ClusterID(customObject)
	return fmt.Sprintf("alias/%s", clusterID)
}

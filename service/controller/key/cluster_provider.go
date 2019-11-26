package key

import (
	"encoding/json"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
)

func clusterProviderSpec(cluster cmainfrastructurev1alpha2.AWSCluster) infrastructurev1alpha2.AWSClusterSpec {
	return mustG8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
}

func clusterProviderStatus(cluster cmainfrastructurev1alpha2.AWSCluster) infrastructurev1alpha2.AWSClusterStatus {
	return mustG8sClusterStatusFromCMAClusterStatus(cluster.Status.ProviderStatus)
}

func mustG8sClusterSpecFromCMAClusterSpec(cmaSpec infrastructurev1alpha2.ProviderSpec) infrastructurev1alpha2.AWSClusterSpec {
	if cmaSpec.Value == nil {
		panic("provider spec value must not be empty")
	}

	var g8sSpec infrastructurev1alpha2.AWSClusterSpec
	{
		if len(cmaSpec.Value.Raw) == 0 {
			return g8sSpec
		}

		err := json.Unmarshal(cmaSpec.Value.Raw, &g8sSpec)
		if err != nil {
			panic(err)
		}
	}

	return g8sSpec
}

func mustG8sClusterStatusFromCMAClusterStatus(cmaStatus *runtime.RawExtension) infrastructurev1alpha2.AWSClusterStatus {
	var g8sStatus infrastructurev1alpha2.AWSClusterStatus
	{
		if cmaStatus == nil {
			return g8sStatus
		}

		if len(cmaStatus.Raw) == 0 {
			return g8sStatus
		}

		err := json.Unmarshal(cmaStatus.Raw, &g8sStatus)
		if err != nil {
			panic(err)
		}
	}

	return g8sStatus
}

package legacykey

import (
	"encoding/json"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func mustG8sSpecFromCMASpec(cmaSpec cmav1alpha1.ProviderSpec) g8sv1alpha1.AWSClusterSpec {
	if cmaSpec.Value == nil {
		panic("provider spec value must not be empty")
	}

	var g8sSpec g8sv1alpha1.AWSClusterSpec
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

func mustG8sStatusFromCMAStatus(cmaStatus *runtime.RawExtension) g8sv1alpha1.AWSClusterStatus {
	if cmaStatus == nil {
		panic("provider status value must not be empty")
	}

	var g8sStatus g8sv1alpha1.AWSClusterStatus
	{
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

func providerSpec(cluster cmav1alpha1.Cluster) g8sv1alpha1.AWSClusterSpec {
	return mustG8sSpecFromCMASpec(cluster.Spec.ProviderSpec)
}

func providerStatus(cluster cmav1alpha1.Cluster) g8sv1alpha1.AWSClusterStatus {
	return mustG8sStatusFromCMAStatus(cluster.Status.ProviderStatus)
}

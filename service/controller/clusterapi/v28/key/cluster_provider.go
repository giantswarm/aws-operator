package key

import (
	"encoding/json"
	"fmt"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func clusterProviderSpec(cluster cmav1alpha1.Cluster) g8sv1alpha1.AWSClusterSpec {
	return mustG8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
}

func clusterProviderStatus(cluster cmav1alpha1.Cluster) g8sv1alpha1.AWSClusterStatus {
	return mustG8sClusterStatusFromCMAClusterStatus(cluster.Status.ProviderStatus)
}

func mustG8sClusterSpecFromCMAClusterSpec(cmaSpec cmav1alpha1.ProviderSpec) g8sv1alpha1.AWSClusterSpec {
	if cmaSpec.Value == nil {
		panic("provider spec value must not be empty")
	}

	var g8sSpec g8sv1alpha1.AWSClusterSpec
	{
		if cmaSpec.Value == nil || len(cmaSpec.Value.Raw) == 0 {
			fmt.Printf("\n")
			fmt.Printf("5\n")
			fmt.Printf("\n")
			return g8sSpec
		}

		b, err := cmaSpec.Value.MarshalJSON()
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(b, &g8sSpec)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("\n")
	fmt.Printf("6\n")
	fmt.Printf("%#v\n", g8sSpec)
	fmt.Printf("\n")

	return g8sSpec
}

func mustG8sClusterStatusFromCMAClusterStatus(cmaStatus *runtime.RawExtension) g8sv1alpha1.AWSClusterStatus {
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

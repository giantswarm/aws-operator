package key

import (
	"encoding/json"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
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

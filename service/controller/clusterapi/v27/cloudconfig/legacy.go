package cloudconfig

import (
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

func cmaClusterToG8sCluster(cr cmav1alpha1.Cluster) g8sv1alpha1.Cluster {
	return g8sv1alpha1.Cluster{
		// TODO
	}
}

func cmaClusterToG8sConfig(cr cmav1alpha1.Cluster) g8sv1alpha1.AWSConfigSpec {
	return g8sv1alpha1.AWSConfigSpec{
		// TODO
	}
}

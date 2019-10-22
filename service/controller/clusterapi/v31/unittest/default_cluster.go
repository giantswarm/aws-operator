package unittest

import (
	"encoding/json"
	"net"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func DefaultCluster() cmav1alpha1.Cluster {
	cr := cmav1alpha1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         "8y5ck",
				label.OperatorVersion: "7.3.0",
			},
		},
	}

	spec := g8sv1alpha1.AWSClusterSpec{
		Cluster: g8sv1alpha1.AWSClusterSpecCluster{
			Description: "Test cluster for template rendering unit test.",
			DNS: g8sv1alpha1.AWSClusterSpecClusterDNS{
				Domain: "gauss.eu-central-1.aws.gigantic.io",
			},
		},
		Provider: g8sv1alpha1.AWSClusterSpecProvider{
			CredentialSecret: g8sv1alpha1.AWSClusterSpecProviderCredentialSecret{
				Name:      "default-credential-secret",
				Namespace: "default",
			},
			Master: g8sv1alpha1.AWSClusterSpecProviderMaster{
				AvailabilityZone: "eu-central-1b",
				InstanceType:     "m5.xlarge",
			},
			Region: "eu-central-1",
		},
	}

	return mustCMAClusterWithG8sProviderSpec(cr, spec)

}

func ClusterWithAZ(cluster cmav1alpha1.Cluster, az string) cmav1alpha1.Cluster {
	region := az[0 : len(az)-1]

	spec := mustG8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
	spec.Provider.Master.AvailabilityZone = az
	spec.Provider.Region = region

	return mustCMAClusterWithG8sProviderSpec(cluster, spec)
}

func ClusterWithNetworkCIDR(cluster cmav1alpha1.Cluster, cidr *net.IPNet) cmav1alpha1.Cluster {
	status := mustG8sClusterStatusFromCMAClusterStatus(cluster.Status.ProviderStatus)

	status.Provider.Network.CIDR = cidr.String()

	return mustCMAClusterWithG8sProviderStatus(cluster, status)
}

func mustG8sClusterSpecFromCMAClusterSpec(cmaSpec cmav1alpha1.ProviderSpec) g8sv1alpha1.AWSClusterSpec {
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

func mustG8sClusterStatusFromCMAClusterStatus(cmaStatus *runtime.RawExtension) g8sv1alpha1.AWSClusterStatus {
	var g8sStatus g8sv1alpha1.AWSClusterStatus
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

func mustCMAClusterWithG8sProviderSpec(cr cmav1alpha1.Cluster, providerExtension g8sv1alpha1.AWSClusterSpec) cmav1alpha1.Cluster {
	var err error

	if cr.Spec.ProviderSpec.Value == nil {
		cr.Spec.ProviderSpec.Value = &runtime.RawExtension{}
	}

	cr.Spec.ProviderSpec.Value.Raw, err = json.Marshal(&providerExtension)
	if err != nil {
		panic(err)
	}

	return cr
}

func mustCMAClusterWithG8sProviderStatus(cr cmav1alpha1.Cluster, providerStatus g8sv1alpha1.AWSClusterStatus) cmav1alpha1.Cluster {
	var err error

	if cr.Status.ProviderStatus == nil {
		cr.Status.ProviderStatus = &runtime.RawExtension{}
	}

	cr.Status.ProviderStatus.Raw, err = json.Marshal(&providerStatus)
	if err != nil {
		panic(err)
	}

	return cr
}

package unittest

import (
	"encoding/json"
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func DefaultCluster() infrastructurev1alpha2.Cluster {
	cr := cmainfrastructurev1alpha2.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         "8y5ck",
				label.OperatorVersion: "7.3.0",
			},
		},
	}

	spec := infrastructurev1alpha2.AWSClusterSpec{
		Cluster: infrastructurev1alpha2.AWSClusterSpecCluster{
			Description: "Test cluster for template rendering unit test.",
			DNS: infrastructurev1alpha2.AWSClusterSpecClusterDNS{
				Domain: "gauss.eu-central-1.aws.gigantic.io",
			},
		},
		Provider: infrastructurev1alpha2.AWSClusterSpecProvider{
			CredentialSecret: infrastructurev1alpha2.AWSClusterSpecProviderCredentialSecret{
				Name:      "default-credential-secret",
				Namespace: "default",
			},
			Master: infrastructurev1alpha2.AWSClusterSpecProviderMaster{
				AvailabilityZone: "eu-central-1b",
				InstanceType:     "m5.xlarge",
			},
			Region: "eu-central-1",
		},
	}

	return mustCMAClusterWithG8sProviderSpec(cr, spec)

}

func ClusterWithAZ(cluster cmainfrastructurev1alpha2.Cluster, az string) cmainfrastructurev1alpha2.Cluster {
	region := az[0 : len(az)-1]

	spec := mustG8sClusterSpecFromCMAClusterSpec(cluster.Spec.ProviderSpec)
	spec.Provider.Master.AvailabilityZone = az
	spec.Provider.Region = region

	return mustCMAClusterWithG8sProviderSpec(cluster, spec)
}

func ClusterWithNetworkCIDR(cluster cmainfrastructurev1alpha2.Cluster, cidr *net.IPNet) cmainfrastructurev1alpha2.Cluster {
	status := mustG8sClusterStatusFromCMAClusterStatus(cluster.Status.ProviderStatus)

	status.Provider.Network.CIDR = cidr.String()

	return mustCMAClusterWithG8sProviderStatus(cluster, status)
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

func mustCMAClusterWithG8sProviderSpec(cr cmainfrastructurev1alpha2.Cluster, providerExtension infrastructurev1alpha2.AWSClusterSpec) cmainfrastructurev1alpha2.Cluster {
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

func mustCMAClusterWithG8sProviderStatus(cr cmainfrastructurev1alpha2.Cluster, providerStatus infrastructurev1alpha2.AWSClusterStatus) cmainfrastructurev1alpha2.Cluster {
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

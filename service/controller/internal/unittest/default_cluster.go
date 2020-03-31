package unittest

import (
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/pkg/label"
)

func ChinaCluster() infrastructurev1alpha2.AWSCluster {
	cr := infrastructurev1alpha2.AWSCluster{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         "8y5ck",
				label.OperatorVersion: "7.3.0",
			},
		},
		Spec: infrastructurev1alpha2.AWSClusterSpec{
			Cluster: infrastructurev1alpha2.AWSClusterSpecCluster{
				Description: "Test china cluster for template rendering unit test.",
				DNS: infrastructurev1alpha2.AWSClusterSpecClusterDNS{
					Domain: "gauss.cn-north-1.aws.gigantic.io",
				},
			},
			Provider: infrastructurev1alpha2.AWSClusterSpecProvider{
				CredentialSecret: infrastructurev1alpha2.AWSClusterSpecProviderCredentialSecret{
					Name:      "default-credential-secret",
					Namespace: "default",
				},
				Master: infrastructurev1alpha2.AWSClusterSpecProviderMaster{
					AvailabilityZone: "cn-north-1a",
					InstanceType:     "m5.xlarge",
				},
				Region: "cn-north-1",
			},
		},
		Status: infrastructurev1alpha2.AWSClusterStatus{
			Provider: infrastructurev1alpha2.AWSClusterStatusProvider{
				Network: infrastructurev1alpha2.AWSClusterStatusProviderNetwork{
					CIDR: "10.0.0.0/24",
				},
			},
		},
	}

	return cr
}

func DefaultCluster() infrastructurev1alpha2.AWSCluster {
	cr := infrastructurev1alpha2.AWSCluster{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         "8y5ck",
				label.OperatorVersion: "7.3.0",
			},
		},
		Spec: infrastructurev1alpha2.AWSClusterSpec{
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
		},
		Status: infrastructurev1alpha2.AWSClusterStatus{
			Provider: infrastructurev1alpha2.AWSClusterStatusProvider{
				Network: infrastructurev1alpha2.AWSClusterStatusProviderNetwork{
					CIDR: "10.0.0.0/24",
				},
			},
		},
	}

	return cr
}

func ClusterWithAZ(cluster infrastructurev1alpha2.AWSCluster, az string) infrastructurev1alpha2.AWSCluster {
	region := az[0 : len(az)-1]

	cluster.Spec.Provider.Master.AvailabilityZone = az
	cluster.Spec.Provider.Region = region

	return cluster
}

func ClusterWithNetworkCIDR(cluster infrastructurev1alpha2.AWSCluster, cidr *net.IPNet) infrastructurev1alpha2.AWSCluster {
	cluster.Status.Provider.Network.CIDR = cidr.String()

	return cluster
}

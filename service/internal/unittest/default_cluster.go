package unittest

import (
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v2/pkg/apis/infrastructure/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"

	"github.com/giantswarm/aws-operator/pkg/label"
)

const (
	DefaultClusterID = "8y5ck"
)

func ChinaCluster() infrastructurev1alpha2.AWSCluster {
	cr := infrastructurev1alpha2.AWSCluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.OperatorVersion: "7.3.0",
				label.Release:         "100.0.0",
			},
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
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
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				label.Cluster:         DefaultClusterID,
				label.OperatorVersion: "7.3.0",
				label.Release:         "100.0.0",
			},
			Name:      DefaultClusterID,
			Namespace: metav1.NamespaceDefault,
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

func DefaultCAPIClusterWithLabels(clusterID string, labels map[string]string) apiv1alpha2.Cluster {
	labels[label.Cluster] = clusterID
	cr := apiv1alpha2.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Labels:    labels,
			Name:      clusterID,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: apiv1alpha2.ClusterSpec{},
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

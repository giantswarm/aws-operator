package unittest

import (
	"encoding/json"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/cluster/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
				Domain: "guux.eu-central-1.aws.gigantic.io",
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

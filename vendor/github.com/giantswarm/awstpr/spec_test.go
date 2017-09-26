package awstpr

import (
	"io/ioutil"
	"net"
	"testing"

	yaml "gopkg.in/yaml.v2"

	"github.com/giantswarm/awstpr/spec"
	"github.com/giantswarm/awstpr/spec/aws"
	"github.com/giantswarm/clustertpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	clustertprdocker "github.com/giantswarm/clustertpr/spec/docker"
	clustertprkubernetes "github.com/giantswarm/clustertpr/spec/kubernetes"
	clustertprkuberneteshyperkube "github.com/giantswarm/clustertpr/spec/kubernetes/hyperkube"
	clustertprkubernetesingress "github.com/giantswarm/clustertpr/spec/kubernetes/ingress"
	clustertprkuberneteskubectl "github.com/giantswarm/clustertpr/spec/kubernetes/kubectl"
	clustertprkubernetesnetworksetup "github.com/giantswarm/clustertpr/spec/kubernetes/networksetup"
	clustertprkubernetesssh "github.com/giantswarm/clustertpr/spec/kubernetes/ssh"
	"github.com/kylelemons/godebug/pretty"
	"github.com/stretchr/testify/require"
)

func TestSpecYamlEncoding(t *testing.T) {
	spec := Spec{
		Cluster: clustertpr.Spec{
			Calico: clustertprspec.Calico{
				CIDR:   16,
				Domain: "giantswarm.io",
				MTU:    1500,
				Subnet: "10.1.2.3",
			},
			Cluster: clustertprspec.Cluster{
				ID: "abc12",
			},
			Customer: clustertprspec.Customer{
				ID: "BooYa",
			},
			Docker: clustertprspec.Docker{
				Daemon: clustertprdocker.Daemon{
					CIDR:      "16",
					ExtraArgs: "--log-opt max-file=1",
				},
			},
			Etcd: clustertprspec.Etcd{
				AltNames: "",
				Domain:   "etcd.giantswarm.io",
				Port:     2379,
				Prefix:   "giantswarm.io",
			},
			Kubernetes: clustertprspec.Kubernetes{
				API: clustertprkubernetes.API{
					AltNames:       "kubernetes,kubernetes.default",
					ClusterIPRange: "172.31.0.0/24",
					Domain:         "api.giantswarm.io",
					IP:             net.ParseIP("172.31.0.1"),
					InsecurePort:   8080,
					SecurePort:     443,
				},
				CloudProvider: "aws",
				DNS: clustertprkubernetes.DNS{
					IP: net.ParseIP("172.31.0.10"),
				},
				Domain: "cluster.giantswarm.io",
				Hyperkube: clustertprkubernetes.Hyperkube{
					Docker: clustertprkuberneteshyperkube.Docker{
						Image: "quay.io/giantswarm/hyperkube",
					},
				},
				IngressController: clustertprkubernetes.IngressController{
					Docker: clustertprkubernetesingress.Docker{
						Image: "quay.io/giantswarm/nginx-ingress-controller",
					},
					Domain:         "ingress.giantswarm.io",
					WildcardDomain: "*.giantswarm.io",
					InsecurePort:   30010,
					SecurePort:     30011,
				},
				Kubectl: clustertprkubernetes.Kubectl{
					Docker: clustertprkuberneteskubectl.Docker{
						Image: "quay.io/giantswarm/docker-kubectl",
					},
				},
				Kubelet: clustertprkubernetes.Kubelet{
					AltNames: "kubernetes,kubernetes.default,kubernetes.default.svc",
					Domain:   "worker.giantswarm.io",
					Labels:   "etcd.giantswarm.io",
					Port:     10250,
				},
				NetworkSetup: clustertprkubernetes.NetworkSetup{
					clustertprkubernetesnetworksetup.Docker{
						Image: "quay.io/giantswarm/k8s-setup-network-environment",
					},
				},
				SSH: clustertprkubernetes.SSH{
					UserList: []clustertprkubernetesssh.User{
						{
							Name:      "xh3b4sd",
							PublicKey: "ssh-rsa AAAAB3NzaC1yc",
						},
					},
				},
			},
			Masters: []clustertprspec.Node{
				{
					ID: "fyz88",
				},
			},
			Vault: clustertprspec.Vault{
				Address: "vault.giantswarm.io",
				Token:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
			},
			Version: "0.1.0",
			Workers: []clustertprspec.Node{
				{
					ID: "axx99",
				},
				{
					ID: "cdd88",
				},
			},
		},
		AWS: spec.AWS{
			Region: "eu-central-1",
			AZ:     "eu-central-1a",
			VPC: aws.VPC{
				CIDR:              "10.0.0.0/16",
				PrivateSubnetCIDR: "10.0.0.0/19",
				PublicSubnetCIDR:  "10.0.128.0/20",
				RouteTableNames: []string{
					"cluster_private_0",
					"cluster_private_1",
				},
				PeerID: "xxxxxxxxx",
			},
			HostedZones: aws.HostedZones{
				API:     "xxxxxxxxxxxxxx",
				Etcd:    "yyyyyyyyyyyyyy",
				Ingress: "zzzzzzzzzzzzzz",
			},
			Masters: []aws.Node{
				{
					ImageID:      "ami-d60ad6b9",
					InstanceType: "m3.large",
				},
			},
			Workers: []aws.Node{
				{
					ImageID:      "ami-d60ad6b9",
					InstanceType: "m3.large",
				},
				{
					ImageID:      "ami-d60ad6b9",
					InstanceType: "m3.large",
				},
			},
		},
	}

	var got map[string]interface{}
	{
		bytes, err := yaml.Marshal(&spec)
		require.NoError(t, err, "marshaling spec")
		err = yaml.Unmarshal(bytes, &got)
		require.NoError(t, err, "unmarshaling spec to map")
	}

	var want map[string]interface{}
	{
		bytes, err := ioutil.ReadFile("testdata/spec.yaml")
		require.NoError(t, err)
		err = yaml.Unmarshal(bytes, &want)
		require.NoError(t, err, "unmarshaling fixture to map")
	}

	diff := pretty.Compare(want, got)
	require.Equal(t, "", diff, "diff: (-want +got)\n%s", diff)
}

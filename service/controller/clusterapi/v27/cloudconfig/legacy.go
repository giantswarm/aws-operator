package cloudconfig

import (
	"net"
	"strings"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/key"
)

func (c *CloudConfig) cmaClusterToG8sConfig(cr cmav1alpha1.Cluster) g8sv1alpha1.AWSConfigSpec {
	return g8sv1alpha1.AWSConfigSpec{
		Cluster: g8sv1alpha1.Cluster{
			Calico: g8sv1alpha1.ClusterCalico{
				CIDR:   c.calicoCIDR,
				MTU:    c.calicoMTU,
				Subnet: c.calicoSubnet,
			},
			Docker: g8sv1alpha1.ClusterDocker{
				Daemon: g8sv1alpha1.ClusterDockerDaemon{
					CIDR: c.dockerDaemonCIDR,
				},
			},
			Etcd: g8sv1alpha1.ClusterEtcd{
				Domain: key.ClusterEtcdEndpoint(cr),
				Prefix: key.EtcdPrefix,
			},
			Kubernetes: g8sv1alpha1.ClusterKubernetes{
				API: g8sv1alpha1.ClusterKubernetesAPI{
					ClusterIPRange: c.clusterIPRange,
					Domain:         key.ClusterAPIEndpoint(cr),
					SecurePort:     key.KubernetesSecurePort,
				},
				CloudProvider: key.CloudProvider,
				DNS: g8sv1alpha1.ClusterKubernetesDNS{
					IP: dnsIPFromRange(c.clusterIPRange),
				},
				Kubelet: g8sv1alpha1.ClusterKubernetesKubelet{
					Domain: key.ClusterKubeletEndpoint(cr),
					Labels: key.KubeletLabels(cr),
				},
				NetworkSetup: g8sv1alpha1.ClusterKubernetesNetworkSetup{
					Docker: g8sv1alpha1.ClusterKubernetesNetworkSetupDocker{
						Image: c.networkSetupDockerImage,
					},
				},
				SSH: g8sv1alpha1.ClusterKubernetesSSH{
					UserList: stringToUserList(c.sshUserList),
				},
			},
		},
		AWS: g8sv1alpha1.AWSConfigSpecAWS{
			Region: key.Region(cr),
		},
	}
}

func dnsIPFromRange(s string) net.IP {
}

func stringToUserList(s string) []g8sv1alpha1.ClusterKubernetesSSHUser {
	var list []g8sv1alpha1.ClusterKubernetesSSHUser

	for _, user := range strings.Split(s, ",") {
		if user == "" {
			continue
		}

		trimmed := strings.TrimSpace(user)
		split := strings.Split(trimmed, ":")

		if len(split) != 2 {
			panic("SSH user format must be <name>:<public key>")
		}

		u := g8sv1alpha1.ClusterKubernetesSSHUser{
			Name:      split[0],
			PublicKey: split[1],
		}

		list = append(list, u)
	}

	return list
}

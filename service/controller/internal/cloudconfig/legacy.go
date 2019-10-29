package cloudconfig

import (
	"net"
	"strings"

	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func cmaClusterToG8sConfig(c Config, cr cmav1alpha1.Cluster, l string) g8sv1alpha1.AWSConfigSpec {
	return g8sv1alpha1.AWSConfigSpec{
		Cluster: g8sv1alpha1.Cluster{
			Calico: g8sv1alpha1.ClusterCalico{
				CIDR:   c.CalicoCIDR,
				MTU:    c.CalicoMTU,
				Subnet: c.CalicoSubnet,
			},
			Docker: g8sv1alpha1.ClusterDocker{
				Daemon: g8sv1alpha1.ClusterDockerDaemon{
					CIDR: c.DockerDaemonCIDR,
				},
			},
			Etcd: g8sv1alpha1.ClusterEtcd{
				Domain: key.ClusterEtcdEndpoint(cr),
				Prefix: key.EtcdPrefix,
			},
			Kubernetes: g8sv1alpha1.ClusterKubernetes{
				API: g8sv1alpha1.ClusterKubernetesAPI{
					ClusterIPRange: c.ClusterIPRange,
					Domain:         key.ClusterAPIEndpoint(cr),
					SecurePort:     key.KubernetesSecurePort,
				},
				CloudProvider: key.CloudProvider,
				DNS: g8sv1alpha1.ClusterKubernetesDNS{
					IP: dnsIPFromRange(c.ClusterIPRange),
				},
				Domain: "cluster.local",
				Kubelet: g8sv1alpha1.ClusterKubernetesKubelet{
					Domain: key.ClusterKubeletEndpoint(cr),
					Labels: l,
				},
				NetworkSetup: g8sv1alpha1.ClusterKubernetesNetworkSetup{
					Docker: g8sv1alpha1.ClusterKubernetesNetworkSetupDocker{
						Image: c.NetworkSetupDockerImage,
					},
				},
				SSH: g8sv1alpha1.ClusterKubernetesSSH{
					UserList: stringToUserList(c.SSHUserList),
				},
			},
		},
		AWS: g8sv1alpha1.AWSConfigSpecAWS{
			Region: key.Region(cr),
		},
	}
}

// dnsIPFromRange takes the cluster IP range and returns the Kube DNS IP we use
// internally. It must be some specific IP, so we chose the last IP octet to be
// 10. The only reason is to do this is to have some static value we apply
// everywhere.
func dnsIPFromRange(s string) net.IP {
	ip := ipFromString(s)
	ip[3] = 10
	return ip
}

func ipFromString(cidr string) net.IP {
	ip, _, err := net.ParseCIDR(cidr)
	if err != nil {
		panic(err)
	}

	// Only IPV4 CIDRs are supported.
	ip = ip.To4()
	if ip == nil {
		panic("CIDR must be an IPV4 range")
	}

	// IP must be a network address.
	if ip[3] != 0 {
		panic("CIDR address must be a network address")
	}

	return ip
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

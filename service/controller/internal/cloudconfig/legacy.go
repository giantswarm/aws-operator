package cloudconfig

import (
	"net"
	"strings"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	g8sv1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/key"
)

func cmaClusterToG8sConfig(c Config, cr infrastructurev1alpha2.Cluster, l string) infrastructurev1alpha2.AWSConfigSpec {
	return infrastructurev1alpha2.AWSConfigSpec{
		Cluster: g8sinfrastructurev1alpha2.Cluster{
			Calico: g8sinfrastructurev1alpha2.ClusterCalico{
				CIDR:   c.CalicoCIDR,
				MTU:    c.CalicoMTU,
				Subnet: c.CalicoSubnet,
			},
			Docker: g8sinfrastructurev1alpha2.ClusterDocker{
				Daemon: g8sinfrastructurev1alpha2.ClusterDockerDaemon{
					CIDR: c.DockerDaemonCIDR,
				},
			},
			Etcd: g8sinfrastructurev1alpha2.ClusterEtcd{
				Domain: key.ClusterEtcdEndpoint(cr),
				Prefix: key.EtcdPrefix,
			},
			Kubernetes: g8sinfrastructurev1alpha2.ClusterKubernetes{
				API: g8sinfrastructurev1alpha2.ClusterKubernetesAPI{
					ClusterIPRange: c.ClusterIPRange,
					Domain:         key.ClusterAPIEndpoint(cr),
					SecurePort:     key.KubernetesSecurePort,
				},
				CloudProvider: key.CloudProvider,
				DNS: g8sinfrastructurev1alpha2.ClusterKubernetesDNS{
					IP: dnsIPFromRange(c.ClusterIPRange),
				},
				Domain: "cluster.local",
				Kubelet: g8sinfrastructurev1alpha2.ClusterKubernetesKubelet{
					Domain: key.ClusterKubeletEndpoint(cr),
					Labels: l,
				},
				NetworkSetup: g8sinfrastructurev1alpha2.ClusterKubernetesNetworkSetup{
					Docker: g8sinfrastructurev1alpha2.ClusterKubernetesNetworkSetupDocker{
						Image: c.NetworkSetupDockerImage,
					},
				},
				SSH: g8sinfrastructurev1alpha2.ClusterKubernetesSSH{
					UserList: stringToUserList(c.SSHUserList),
				},
			},
		},
		AWS: infrastructurev1alpha2.AWSConfigSpecAWS{
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

func stringToUserList(s string) []g8sinfrastructurev1alpha2.ClusterKubernetesSSHUser {
	var list []g8sinfrastructurev1alpha2.ClusterKubernetesSSHUser

	for _, user := range strings.Split(s, ",") {
		if user == "" {
			continue
		}

		trimmed := strings.TrimSpace(user)
		split := strings.Split(trimmed, ":")

		if len(split) != 2 {
			panic("SSH user format must be <name>:<public key>")
		}

		u := g8sinfrastructurev1alpha2.ClusterKubernetesSSHUser{
			Name:      split[0],
			PublicKey: split[1],
		}

		list = append(list, u)
	}

	return list
}

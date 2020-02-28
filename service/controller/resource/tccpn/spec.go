package tccpn

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
)

const (
	cadvisorPort         = 4194
	etcdPort             = 2379
	kubeletPort          = 10250
	nodeExporterPort     = 10300
	kubeStateMetricsPort = 10301
	sshPort              = 22

	tcpProtocol = "tcp"

	defaultCIDR = "0.0.0.0/0"
)

// APIWhitelist defines guest cluster k8s public/private api whitelisting.
type APIWhitelist struct {
	Private Whitelist
	Public  Whitelist
}

// Whitelist represents the structure required for defining whitelisting for
// resource security group
type Whitelist struct {
	Enabled    bool
	SubnetList string
}

type securityConfig struct {
	APIWhitelist                    APIWhitelist
	ControlPlaneNATGatewayAddresses []*ec2.Address
	ControlPlaneVPCCidr             string
	ProviderCIDR                    string
	CustomObject                    infrastructurev1alpha2.AWSControlPlane
}

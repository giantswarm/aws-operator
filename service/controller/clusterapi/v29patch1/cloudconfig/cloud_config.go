package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/encrypter"
)

const (
	FileOwnerUserName  = "root"
	FileOwnerGroupName = "root"
	FilePermission     = 0700
)

// Config represents the configuration used to create a cloud config service.
type Config struct {
	Encrypter encrypter.Interface
	Logger    micrologger.Logger

	CalicoCIDR                int
	CalicoMTU                 int
	CalicoSubnet              string
	ClusterIPRange            string
	DockerDaemonCIDR          string
	IgnitionPath              string
	ImagePullProgressDeadline string
	NetworkSetupDockerImage   string
	OIDC                      ConfigOIDC
	PodInfraContainerImage    string
	RegistryDomain            string
	SSHUserList               string
	SSOPublicKey              string
}

type ConfigOIDC struct {
	ClientID      string
	IssuerURL     string
	UsernameClaim string
	GroupsClaim   string
}

// CloudConfig implements the cloud config service interface.
type CloudConfig struct {
	encrypter encrypter.Interface
	logger    micrologger.Logger

	k8sAPIExtraArgs     []string
	k8sKubeletExtraArgs []string

	calicoCIDR                int
	calicoMTU                 int
	calicoSubnet              string
	clusterIPRange            string
	dockerDaemonCIDR          string
	ignitionPath              string
	imagePullProgressDeadline string
	networkSetupDockerImage   string
	registryDomain            string
	sshUserList               string
	ssoPublicKey              string
}

// New creates a new configured cloud config service.
func New(config Config) (*CloudConfig, error) {
	if config.Encrypter == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", config)
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	var k8sAPIExtraArgs []string
	{
		if config.OIDC.ClientID != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-client-id=%s", config.OIDC.ClientID))
		}
		if config.OIDC.IssuerURL != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", config.OIDC.IssuerURL))
		}
		if config.OIDC.UsernameClaim != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", config.OIDC.UsernameClaim))
		}
		if config.OIDC.GroupsClaim != "" {
			k8sAPIExtraArgs = append(k8sAPIExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", config.OIDC.GroupsClaim))
		}
	}

	var k8sKubeletExtraArgs []string
	{
		if config.PodInfraContainerImage != "" {
			k8sKubeletExtraArgs = append(k8sKubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", config.PodInfraContainerImage))
		}
	}

	if config.CalicoCIDR == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.CalicoCIDR must not be empty", config)
	}
	if config.CalicoMTU == 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.CalicoMTU must not be empty", config)
	}
	if config.CalicoSubnet == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.CalicoSubnet must not be empty", config)
	}
	if config.ClusterIPRange == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", config)
	}
	if config.DockerDaemonCIDR == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.DockerDaemonCIDR must not be empty", config)
	}
	if config.IgnitionPath == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", config)
	}
	if config.ImagePullProgressDeadline == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.ImagePullProgressDeadline must not be empty", config)
	}
	if config.NetworkSetupDockerImage == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.NetworkSetupDockerImage must not be empty", config)
	}
	if config.RegistryDomain == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", config)
	}
	if config.SSHUserList == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.SSHUserList must not be empty", config)
	}
	if config.SSOPublicKey == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.SSOPublicKey must not be empty", config)
	}

	newCloudConfig := &CloudConfig{
		encrypter: config.Encrypter,
		logger:    config.Logger,

		k8sAPIExtraArgs:     k8sAPIExtraArgs,
		k8sKubeletExtraArgs: k8sKubeletExtraArgs,

		calicoCIDR:                config.CalicoCIDR,
		calicoMTU:                 config.CalicoMTU,
		calicoSubnet:              config.CalicoSubnet,
		clusterIPRange:            config.ClusterIPRange,
		dockerDaemonCIDR:          config.DockerDaemonCIDR,
		ignitionPath:              config.IgnitionPath,
		imagePullProgressDeadline: config.ImagePullProgressDeadline,
		networkSetupDockerImage:   config.NetworkSetupDockerImage,
		registryDomain:            config.RegistryDomain,
		sshUserList:               config.SSHUserList,
		ssoPublicKey:              config.SSOPublicKey,
	}

	return newCloudConfig, nil
}

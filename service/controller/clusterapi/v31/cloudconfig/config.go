package cloudconfig

import (
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v31/encrypter"
)

type Config struct {
	Encrypter encrypter.Interface
	Logger    micrologger.Logger

	APIExtraArgs              []string
	CalicoCIDR                int
	CalicoMTU                 int
	CalicoSubnet              string
	ClusterIPRange            string
	DockerDaemonCIDR          string
	IgnitionPath              string
	ImagePullProgressDeadline string
	KubeletExtraArgs          []string
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

func (c Config) Default() Config {
	if c.OIDC.ClientID != "" {
		c.APIExtraArgs = append(c.APIExtraArgs, fmt.Sprintf("--oidc-client-id=%s", c.OIDC.ClientID))
	}
	if c.OIDC.IssuerURL != "" {
		c.APIExtraArgs = append(c.APIExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", c.OIDC.IssuerURL))
	}
	if c.OIDC.UsernameClaim != "" {
		c.APIExtraArgs = append(c.APIExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", c.OIDC.UsernameClaim))
	}
	if c.OIDC.GroupsClaim != "" {
		c.APIExtraArgs = append(c.APIExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", c.OIDC.GroupsClaim))
	}

	if c.PodInfraContainerImage != "" {
		c.KubeletExtraArgs = append(c.KubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", c.PodInfraContainerImage))
	}

	return c
}

func (c Config) Validate() error {
	if c.Encrypter == nil {
		return microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", c)
	}
	if c.Logger == nil {
		return microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}

	if c.CalicoCIDR == 0 {
		return microerror.Maskf(invalidConfigError, "%T.CalicoCIDR must not be empty", c)
	}
	if c.CalicoMTU == 0 {
		return microerror.Maskf(invalidConfigError, "%T.CalicoMTU must not be empty", c)
	}
	if c.CalicoSubnet == "" {
		return microerror.Maskf(invalidConfigError, "%T.CalicoSubnet must not be empty", c)
	}
	if c.ClusterIPRange == "" {
		return microerror.Maskf(invalidConfigError, "%T.ClusterIPRange must not be empty", c)
	}
	if c.DockerDaemonCIDR == "" {
		return microerror.Maskf(invalidConfigError, "%T.DockerDaemonCIDR must not be empty", c)
	}
	if c.IgnitionPath == "" {
		return microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", c)
	}
	if c.ImagePullProgressDeadline == "" {
		return microerror.Maskf(invalidConfigError, "%T.ImagePullProgressDeadline must not be empty", c)
	}
	if c.NetworkSetupDockerImage == "" {
		return microerror.Maskf(invalidConfigError, "%T.NetworkSetupDockerImage must not be empty", c)
	}
	if c.RegistryDomain == "" {
		return microerror.Maskf(invalidConfigError, "%T.RegistryDomain must not be empty", c)
	}
	if c.SSHUserList == "" {
		return microerror.Maskf(invalidConfigError, "%T.SSHUserList must not be empty", c)
	}
	if c.SSOPublicKey == "" {
		return microerror.Maskf(invalidConfigError, "%T.SSOPublicKey must not be empty", c)
	}

	return nil
}

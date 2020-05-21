package cloudconfig

import (
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"

	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
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
	ExternalSNAT              bool
	IgnitionPath              string
	ImagePullProgressDeadline string
	KubeletExtraArgs          []string
	ClusterDomain             string
	NetworkSetupDockerImage   string
	PodInfraContainerImage    string
	RegistryDomain            string
	SSHUserList               string
	SSOPublicKey              string
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
	if c.ClusterDomain == "" {
		return microerror.Maskf(invalidConfigError, "%T.ClusterDomain must not be empty", c)
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

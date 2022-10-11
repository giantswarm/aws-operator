package cloudconfig

import (
	"github.com/giantswarm/certs/v4/pkg/certs"
	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/randomkeys/v3"

	"github.com/giantswarm/aws-operator/v14/service/internal/cloudtags"
	"github.com/giantswarm/aws-operator/v14/service/internal/encrypter"
	"github.com/giantswarm/aws-operator/v14/service/internal/hamaster"
	"github.com/giantswarm/aws-operator/v14/service/internal/images"
	event "github.com/giantswarm/aws-operator/v14/service/internal/recorder"
)

type Config struct {
	CertsSearcher      certs.Interface
	CloudTags          cloudtags.Interface
	Encrypter          encrypter.Interface
	Event              event.Interface
	HAMaster           hamaster.Interface
	Images             images.Interface
	K8sClient          k8sclient.Interface
	Logger             micrologger.Logger
	RandomKeysSearcher randomkeys.Interface

	APIExtraArgs            []string
	CalicoCIDR              int
	CalicoMTU               int
	CalicoSubnet            string
	ClusterIPRange          string
	DockerDaemonCIDR        string
	DockerhubToken          string
	ExternalSNAT            bool
	IgnitionPath            string
	KubeletExtraArgs        []string
	ClusterDomain           string
	NetworkSetupDockerImage string
	PodInfraContainerImage  string
	RegistryDomain          string
	RegistryMirrors         []string
	SSHUserList             string
	SSOPublicKey            string
}

func (c Config) Validate() error {
	if c.CertsSearcher == nil {
		return microerror.Maskf(invalidConfigError, "%T.CertsSearcher must not be empty", c)
	}
	if c.Encrypter == nil {
		return microerror.Maskf(invalidConfigError, "%T.Encrypter must not be empty", c)
	}
	if c.Event == nil {
		return microerror.Maskf(invalidConfigError, "%T.Event must not be empty", c)
	}
	if c.HAMaster == nil {
		return microerror.Maskf(invalidConfigError, "%T.HAMaster must not be empty", c)
	}
	if c.Images == nil {
		return microerror.Maskf(invalidConfigError, "%T.Images must not be empty", c)
	}
	if c.K8sClient == nil {
		return microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", c)
	}
	if c.Logger == nil {
		return microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", c)
	}
	if c.RandomKeysSearcher == nil {
		return microerror.Maskf(invalidConfigError, "%T.RandomKeysSearcher must not be empty", c)
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
	if c.DockerhubToken == "" {
		return microerror.Maskf(invalidConfigError, "%T.DockerhubToken must not be empty", c)
	}

	if c.IgnitionPath == "" {
		return microerror.Maskf(invalidConfigError, "%T.IgnitionPath must not be empty", c)
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

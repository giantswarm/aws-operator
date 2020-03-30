package cloudconfig

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_5_2_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCCPConfig struct {
	Config Config
}

type TCCP struct {
	config Config
}

func NewTCCP(config TCCPConfig) (*TCCP, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &TCCP{
		config: config.Config,
	}

	return t, nil
}

func (t *TCCP) Render(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, labels string) ([]byte, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	randomKeyTmplSet, err := renderRandomKeyTmplSet(ctx, t.config.Encrypter, cc.Status.TenantCluster.Encryption.Key, clusterKeys)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var apiExtraArgs []string
	{
		if key.OIDCClientID(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-client-id=%s", key.OIDCClientID(cr)))
		}
		if key.OIDCIssuerURL(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", key.OIDCIssuerURL(cr)))
		}
		if key.OIDCUsernameClaim(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", key.OIDCUsernameClaim(cr)))
		}
		if key.OIDCGroupsClaim(cr) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", key.OIDCGroupsClaim(cr)))
		}

		apiExtraArgs = append(apiExtraArgs, t.config.APIExtraArgs...)
	}

	var kubeletExtraArgs []string
	{
		if t.config.PodInfraContainerImage != "" {
			kubeletExtraArgs = append(kubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", t.config.PodInfraContainerImage))
		}

		kubeletExtraArgs = append(kubeletExtraArgs, t.config.KubeletExtraArgs...)
	}

	var params k8scloudconfig.Params
	{
		params = k8scloudconfig.DefaultParams()

		params.Cluster = cmaClusterToG8sConfig(t.config, cr, labels).Cluster
		params.DisableEncryptionAtREST = true
		// Ingress Controller service is not created via ignition.
		// It gets created by the Ingress Controller app if it is installed in the tenant cluster.
		params.DisableIngressControllerService = true
		params.EtcdPort = key.EtcdPort
		params.Extension = &MasterExtension{
			awsConfigSpec: cmaClusterToG8sConfig(t.config, cr, labels),
			baseExtension: baseExtension{
				registryDomain: t.config.RegistryDomain,
				cluster:        cr,
				encrypter:      t.config.Encrypter,
				encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
			},
			cc:               cc,
			clusterCerts:     clusterCerts,
			randomKeyTmplSet: randomKeyTmplSet,
		}
		params.Hyperkube.Apiserver.Pod.CommandExtraArgs = apiExtraArgs
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = kubeletExtraArgs
		params.ImagePullProgressDeadline = t.config.ImagePullProgressDeadline
		params.RegistryDomain = t.config.RegistryDomain
		params.SSOPublicKey = t.config.SSOPublicKey

		ignitionPath := k8scloudconfig.GetIgnitionPath(t.config.IgnitionPath)
		params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var templateBody []byte
	{
		c := k8scloudconfig.CloudConfigConfig{
			Params:   params,
			Template: k8scloudconfig.MasterTemplate,
		}

		cloudConfig, err := k8scloudconfig.NewCloudConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = cloudConfig.ExecuteTemplate()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		templateBody = []byte(cloudConfig.String())
	}

	return templateBody, nil
}

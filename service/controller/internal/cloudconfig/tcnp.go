package cloudconfig

import (
	"context"
	"fmt"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCNPConfig struct {
	Config Config
}

type TCNP struct {
	config Config
}

func NewTCNP(config TCNPConfig) (*TCNP, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &TCNP{
		config: config.Config,
	}

	return t, nil
}

func (t *TCNP) NewPaths(ctx context.Context, obj interface{}) ([]string, error) {
	return nil, nil
}

func (t *TCNP) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var cl infrastructurev1alpha2.AWSCluster
	{
		var list infrastructurev1alpha2.AWSClusterList
		err := t.config.K8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if len(list.Items) != 1 {
			// TODO return package error
		}

		cl = list.Items[0]
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
		// Default registry, kubernetes, etcd images etcd.
		// Required for proper rending of the templates.
		params = k8scloudconfig.DefaultParams()

		params.Cluster = cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCNP(&cr)).Cluster
		params.Extension = &TCNPExtension{
			awsConfigSpec:  cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCNP(&cr)),
			cc:             cc,
			cluster:        cl,
			clusterCerts:   clusterCerts,
			encrypter:      t.config.Encrypter,
			encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
			registryDomain: t.config.RegistryDomain,
		}
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = kubeletExtraArgs
		params.Images = images
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
			Template: k8scloudconfig.WorkerTemplate,
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

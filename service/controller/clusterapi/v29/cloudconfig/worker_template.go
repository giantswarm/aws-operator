package cloudconfig

import (
	"context"

	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_6_0"
	"github.com/giantswarm/microerror"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a string.
func (c *CloudConfig) NewWorkerTemplate(ctx context.Context, cr cmav1alpha1.Cluster, clusterCerts certs.Cluster) (string, error) {
	var err error

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var params k8scloudconfig.Params
	{
		be := baseExtension{
			cluster:       cr,
			encrypter:     c.encrypter,
			encryptionKey: cc.Status.TenantCluster.Encryption.Key,
		}

		// Default registry, kubernetes, etcd images etcd.
		// Required for proper rending of the templates.
		params = k8scloudconfig.DefaultParams()

		params.Cluster = c.cmaClusterToG8sConfig(cr).Cluster
		params.Extension = &WorkerExtension{
			awsConfigSpec: c.cmaClusterToG8sConfig(cr),
			baseExtension: be,
			ctlCtx:        cc,

			ClusterCerts: clusterCerts,
		}
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = c.k8sKubeletExtraArgs
		params.RegistryDomain = c.registryDomain
		params.SSOPublicKey = c.ssoPublicKey

		ignitionPath := k8scloudconfig.GetIgnitionPath(c.ignitionPath)
		params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		cloudConfigConfig := k8scloudconfig.DefaultCloudConfigConfig()
		cloudConfigConfig.Params = params
		cloudConfigConfig.Template = k8scloudconfig.WorkerTemplate

		newCloudConfig, err = k8scloudconfig.NewCloudConfig(cloudConfigConfig)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return newCloudConfig.String(), nil
}

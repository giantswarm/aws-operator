package cloudconfig

import (
	"context"

	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_7_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	cmav1alpha1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29/key"
)

// NewMasterTemplate generates a new master cloud config template and returns it
// as a string.
func (c *CloudConfig) NewMasterTemplate(ctx context.Context, cr cmav1alpha1.Cluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster) (string, error) {
	var err error

	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	randomKeyTmplSet, err := renderRandomKeyTmplSet(ctx, c.encrypter, cc.Status.TenantCluster.Encryption.Key, clusterKeys)
	if err != nil {
		return "", microerror.Mask(err)
	}

	be := baseExtension{
		cluster:       cr,
		encrypter:     c.encrypter,
		encryptionKey: cc.Status.TenantCluster.Encryption.Key,
	}

	masterExtension := &MasterExtension{
		awsConfigSpec: c.cmaClusterToG8sConfig(cr),
		baseExtension: be,
		ctlCtx:        cc,

		ClusterCerts:     clusterCerts,
		RandomKeyTmplSet: randomKeyTmplSet,
	}

	files, err := masterExtension.Files(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}
	units, err := masterExtension.Units()
	if err != nil {
		return "", microerror.Mask(err)
	}

	var params k8scloudconfig.Params
	{

		params = k8scloudconfig.DefaultParams()

		params.Cluster = c.cmaClusterToG8sConfig(cr).Cluster
		params.DisableEncryptionAtREST = true
		// Ingress controller service remains in k8scloudconfig and will be
		// removed in a later migration.
		params.DisableIngressControllerService = false
		params.EtcdPort = key.EtcdPort
		params.Extension = k8scloudconfig.Extension{
			Files: files,
			Units: units,
		}
		params.Hyperkube.Apiserver.Pod.CommandExtraArgs = c.k8sAPIExtraArgs
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = c.k8sKubeletExtraArgs
		params.ImagePullProgressDeadline = c.imagePullProgressDeadline
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
		cloudConfigConfig.Template = k8scloudconfig.MasterTemplate

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

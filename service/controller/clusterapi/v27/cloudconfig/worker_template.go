package cloudconfig

import (
	"context"
	"encoding/base64"

	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_3_0"
	"github.com/giantswarm/microerror"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v27/templates/cloudconfig"
)

// NewWorkerTemplate generates a new worker cloud config template and returns it
// as a string.
func (c *CloudConfig) NewWorkerTemplate(ctx context.Context, cr v1alpha1.Cluster, clusterCerts certs.Cluster) (string, error) {
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

		params.Cluster = cmaClusterToG8sCluster(cr)
		params.Extension = &WorkerExtension{
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

type WorkerExtension struct {
	baseExtension

	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	// See https://github.com/giantswarm/giantswarm/issues/4329.
	//
	ctlCtx *controllercontext.Context

	ClusterCerts certs.Cluster
}

func (e *WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	// See https://github.com/giantswarm/giantswarm/issues/4329.
	//
	ctx := context.TODO()

	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScript,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner: k8scloudconfig.Owner{
				User:  FileOwnerUser,
				Group: FileOwnerGroup,
			},
			Permissions: 0700,
		},
		{
			AssetContent: cloudconfig.WaitDockerConf,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner: k8scloudconfig.Owner{
				User:  FileOwnerUser,
				Group: FileOwnerGroup,
			},
			Permissions: 0700,
		},
	}

	certsMeta := []k8scloudconfig.FileMetadata{}
	{
		certFiles := certs.NewFilesClusterWorker(e.ClusterCerts)

		for _, f := range certFiles {
			// TODO We should just pass ctx to Files.
			//
			// See https://github.com/giantswarm/giantswarm/issues/4329.
			//
			ctx = controllercontext.NewContext(ctx, *e.ctlCtx)

			data, err := e.encrypt(ctx, f.Data)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			meta := k8scloudconfig.FileMetadata{
				AssetContent: string(data),
				Path:         f.AbsolutePath + ".enc",
				Owner: k8scloudconfig.Owner{
					User:  FileOwnerUser,
					Group: FileOwnerGroup,
				},
				Permissions: 0700,
			}

			certsMeta = append(certsMeta, meta)
		}
	}

	var fileAssets []k8scloudconfig.FileAsset

	data := e.templateData()

	for _, m := range filesMeta {
		c, err := k8scloudconfig.RenderFileAssetContent(m.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		asset := k8scloudconfig.FileAsset{
			Metadata: m,
			Content:  c,
		}

		fileAssets = append(fileAssets, asset)
	}

	for _, cm := range certsMeta {
		c := base64.StdEncoding.EncodeToString([]byte(cm.AssetContent))
		asset := k8scloudconfig.FileAsset{
			Metadata: cm,
			Content:  c,
		}

		fileAssets = append(fileAssets, asset)
	}

	return fileAssets, nil
}

func (e *WorkerExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.PersistentVarLibDockerMount,
			Name:         "var-lib-docker.mount",
			Enabled:      true,
		},
		// Set bigger timeouts for NVME driver.
		// Workaround for https://github.com/coreos/bugs/issues/2484
		// TODO issue: https://github.com/giantswarm/giantswarm/issues/4255
		{
			AssetContent: cloudconfig.NVMESetTimeoutsUnit,
			Name:         "nvme-set-timeouts.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.SetHostname,
			Name:         "set-hostname.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.EphemeralVarLogMount,
			Name:         "var-log.mount",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.EphemeralVarLibKubeletMount,
			Name:         "var-lib-kubelet.mount",
			Enabled:      true,
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, m := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(m.AssetContent, cmaClusterToG8sConfig(e.cluster))
		if err != nil {
			return nil, microerror.Mask(err)
		}

		unitAsset := k8scloudconfig.UnitAsset{
			Metadata: m,
			Content:  c,
		}

		newUnits = append(newUnits, unitAsset)
	}

	return newUnits, nil
}

func (e *WorkerExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	newSections := []k8scloudconfig.VerbatimSection{}

	return newSections
}

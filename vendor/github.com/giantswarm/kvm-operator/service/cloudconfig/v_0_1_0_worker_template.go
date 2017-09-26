package cloudconfig

import (
	"github.com/giantswarm/certificatetpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_0_1_0"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
)

func v_0_1_0WorkerTemplate(customObject kvmtpr.CustomObject, certs certificatetpr.AssetsBundle, node clustertprspec.Node) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &v_0_1_0WorkerExtension{
			certs: certs,
		}
		params.Node = node
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		newCloudConfig, err = k8scloudconfig.NewCloudConfig(k8scloudconfig.WorkerTemplate, params)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = newCloudConfig.ExecuteTemplate()
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	return newCloudConfig.Base64(), nil
}

type v_0_1_0WorkerExtension struct {
	certs certificatetpr.AssetsBundle
}

func (e *v_0_1_0WorkerExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		// Calico client.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/calico/client-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/calico/client-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/calico/client-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Etcd client.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/etcd/client-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/etcd/client-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/etcd/client-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Kubernetes worker.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/worker-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/worker-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/worker-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
	}

	var newFiles []k8scloudconfig.FileAsset

	for _, fm := range filesMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		fileAsset := k8scloudconfig.FileAsset{
			Metadata: fm,
			Content:  c,
		}

		newFiles = append(newFiles, fileAsset)
	}

	return newFiles, nil
}

func (e *v_0_1_0WorkerExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	return nil, nil
}

func (e *v_0_1_0WorkerExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}

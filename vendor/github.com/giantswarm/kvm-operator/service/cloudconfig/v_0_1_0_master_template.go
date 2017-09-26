package cloudconfig

import (
	"github.com/giantswarm/certificatetpr"
	clustertprspec "github.com/giantswarm/clustertpr/spec"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_0_1_0"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
)

func v_0_1_0MasterTemplate(customObject kvmtpr.CustomObject, certs certificatetpr.AssetsBundle, node clustertprspec.Node) (string, error) {
	var err error

	var params k8scloudconfig.Params
	{
		params.Cluster = customObject.Spec.Cluster
		params.Extension = &v_0_1_0MasterExtension{
			certs: certs,
		}
		params.Node = node
	}

	var newCloudConfig *k8scloudconfig.CloudConfig
	{
		newCloudConfig, err = k8scloudconfig.NewCloudConfig(k8scloudconfig.MasterTemplate, params)
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

type v_0_1_0MasterExtension struct {
	certs certificatetpr.AssetsBundle
}

func (e *v_0_1_0MasterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	filesMeta := []k8scloudconfig.FileMetadata{
		// Kubernetes API server.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/apiserver-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/apiserver-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/apiserver-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
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
		// Etcd server.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/etcd/server-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/etcd/server-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/etcd/server-key.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		// Service account.
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.CA}]),
			Path:         "/etc/kubernetes/ssl/service-account-ca.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.Crt}]),
			Path:         "/etc/kubernetes/ssl/service-account-crt.pem",
			Owner:        FileOwner,
			Permissions:  FilePermission,
		},
		{
			AssetContent: string(e.certs[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.Key}]),
			Path:         "/etc/kubernetes/ssl/service-account-key.pem",
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

func (e *v_0_1_0MasterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		// Mount etcd volume when directory first accessed
		// This automount is workaround for
		// https://bugzilla.redhat.com/show_bug.cgi?id=1184122
		{
			AssetContent: `[Unit]
Description=Automount for etcd volume
[Automount]
Where=/etc/kubernetes/data/etcd
[Install]
WantedBy=multi-user.target
`,
			Name:    "etc-kubernetes-data-etcd.automount",
			Enable:  true,
			Command: "start",
		},
		// Mount for etcd volume activated by automount
		{
			AssetContent: `[Unit]
Description=Mount for etcd volume
[Mount]
What=etcdshare
Where=/etc/kubernetes/data/etcd
Options=trans=virtio,version=9p2000.L,cache=mmap
Type=9p
[Install]
WantedBy=multi-user.target
`,
			Name:   "etc-kubernetes-data-etcd.mount",
			Enable: false,
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	for _, fm := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, nil)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		unitAsset := k8scloudconfig.UnitAsset{
			Metadata: fm,
			Content:  c,
		}

		newUnits = append(newUnits, unitAsset)
	}

	return newUnits, nil
}

func (e *v_0_1_0MasterExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	return nil
}

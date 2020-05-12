package cloudconfig

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/clientset/versioned"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	cloudconfig "github.com/giantswarm/aws-operator/service/controller/internal/cloudconfig/template"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCCPNConfig struct {
	Config    Config
	G8sClient versioned.Interface
}

type TCCPN struct {
	config    Config
	g8sClient versioned.Interface
}

func NewTCCPN(config TCCPNConfig) (*TCCPN, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if config.G8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T cannot by nil", config.G8sClient)
	}

	t := &TCCPN{
		config:    config.Config,
		g8sClient: config.G8sClient,
	}

	return t, nil
}

func (t *TCCPN) Render(ctx context.Context, cr infrastructurev1alpha2.AWSCluster, clusterCerts certs.Cluster, clusterKeys randomkeys.Cluster, images k8scloudconfig.Images, labels string) ([]byte, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	listOptions := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", key.TagCluster, key.ClusterID(&cr)),
	}

	cps, err := t.g8sClient.InfrastructureV1alpha2().AWSControlPlanes(cr.Namespace).List(listOptions)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if len(cps.Items) != 1 {
		return nil, microerror.Maskf(executionFailedError, "expected 1 control plane cr but found %d", len(cps.Items))
	}
	cp := cps.Items[0]

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

	masterID := 0 // for now we have only 1 master, TODO get this value via render function as argument

	var masterSubnet net.IPNet
	{
		zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones
		for _, az := range zones {
			if az.Name == key.ControlPlaneAvailabilityZones(cp)[masterID] {
				masterSubnet = az.Subnet.Private.CIDR
				break
			}
		}
	}

	// Allow the actual externalSNAT to be set by the CR.
	var externalSNAT bool
	if key.ExternalSNAT(cr) == nil {
		externalSNAT = t.config.ExternalSNAT
	} else {
		externalSNAT = *key.ExternalSNAT(cr)
	}

	var params k8scloudconfig.Params
	{
		params = k8scloudconfig.DefaultParams()

		g8sConfig := cmaClusterToG8sConfig(t.config, cr, labels)
		params.Cluster = g8sConfig.Cluster
		params.DisableEncryptionAtREST = true
		// Ingress Controller service is not created via ignition.
		// It gets created by the Ingress Controller app if it is installed in the tenant cluster.
		params.DisableIngressControllerService = true
		params.EtcdPort = key.EtcdPort
		params.Extension = &HAMasterExtension{
			baseExtension: baseExtension{
				cluster:        cr,
				encrypter:      t.config.Encrypter,
				encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
				externalSNAT:   externalSNAT,
				masterSubnet:   masterSubnet,
				masterID:       masterID,
				registryDomain: t.config.RegistryDomain,
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
		params.Images = images
		params.EnableAWSCNI = true

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

type HAMasterExtension struct {
	baseExtension
	// TODO Pass context to k8scloudconfig rendering fucntions
	//
	//	See https://github.com/giantswarm/giantswarm/issues/4329.
	//
	cc               *controllercontext.Context
	clusterCerts     certs.Cluster
	randomKeyTmplSet RandomKeyTmplSet
}

func (e *HAMasterExtension) Files() ([]k8scloudconfig.FileAsset, error) {
	ctx := context.TODO()

	storageClass := cloudconfig.InstanceStorageClassEncryptedContent

	filesMeta := []k8scloudconfig.FileMetadata{
		{
			AssetContent: cloudconfig.DecryptTLSAssetsScript,
			Path:         "/opt/bin/decrypt-tls-assets",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: FilePermission,
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsScript,
			Path:         "/opt/bin/decrypt-keys-assets",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: FilePermission,
		},

		{
			AssetContent: e.randomKeyTmplSet.APIServerEncryptionKey,
			Path:         "/etc/kubernetes/encryption/k8s-encryption-config.yaml.enc",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: cloudconfig.WaitDockerConf,
			Path:         "/etc/systemd/system/docker.service.d/01-wait-docker.conf",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: FilePermission,
		},
		// Add use-proxy-protocol to ingress-controller ConfigMap, this doesn't work
		// on KVM because of dependencies on hardware LB configuration.
		{
			AssetContent: cloudconfig.IngressControllerConfigMap,
			Path:         "/srv/ingress-controller-cm.yml",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		// NVME disks udev rules and script.
		// Workaround for https://github.com/coreos/bugs/issues/2399
		{
			AssetContent: cloudconfig.NVMEUdevRule,
			Path:         "/etc/udev/rules.d/10-ebs-nvme-mapping.rules",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: cloudconfig.NVMEUdevScript,
			Path:         "/opt/ebs-nvme-mapping",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0766,
		},
		{
			AssetContent: storageClass,
			Path:         "/srv/default-storage-class.yaml",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: cloudconfig.Etcd3ExtraConfig,
			Path:         "/etc/systemd/system/etcd3.d/10-require-attach-dep.conf",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: cloudconfig.SystemdNetworkdEth1Network,
			Path:         "/etc/systemd/network/10-eth1.network",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
		{
			AssetContent: cloudconfig.AwsCNIManifest,
			Path:         "/srv/aws-cni.yaml",
			Owner: k8scloudconfig.Owner{
				Group: k8scloudconfig.Group{
					Name: FileOwnerGroupName,
				},
				User: k8scloudconfig.User{
					Name: FileOwnerUserName,
				},
			},
			Permissions: 0644,
		},
	}

	certsMeta := []k8scloudconfig.FileMetadata{}
	{
		certFiles := certs.NewFilesClusterMaster(e.clusterCerts)

		for _, f := range certFiles {
			// TODO We should just pass ctx to Files.
			//
			// 	See https://github.com/giantswarm/giantswarm/issues/4329.
			//
			ctx = controllercontext.NewContext(ctx, *e.cc)

			data, err := e.encrypt(ctx, f.Data)
			if err != nil {
				return nil, microerror.Mask(err)
			}

			meta := k8scloudconfig.FileMetadata{
				AssetContent: string(data),
				Path:         f.AbsolutePath + ".enc",
				Owner: k8scloudconfig.Owner{
					Group: k8scloudconfig.Group{
						Name: FileOwnerGroupName,
					},
					User: k8scloudconfig.User{
						Name: FileOwnerUserName,
					},
				},
				Permissions: 0700,
			}

			certsMeta = append(certsMeta, meta)
		}
	}

	var fileAssets []k8scloudconfig.FileAsset

	data := e.templateDataTCCPN()

	for _, fm := range filesMeta {
		c, err := k8scloudconfig.RenderFileAssetContent(fm.AssetContent, data)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		asset := k8scloudconfig.FileAsset{
			Metadata: fm,
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

func (e *HAMasterExtension) Units() ([]k8scloudconfig.UnitAsset, error) {
	unitsMeta := []k8scloudconfig.UnitMetadata{
		// Create symlinks for nvme disks.
		// This service should be started only on first boot.
		{
			AssetContent: cloudconfig.NVMEUdevTriggerUnit,
			Name:         "ebs-nvme-udev-trigger.service",
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
			AssetContent: cloudconfig.DecryptTLSAssetsService,
			Name:         "decrypt-tls-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.DecryptKeysAssetsService,
			Name:         "decrypt-keys-assets.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.SetHostname,
			Name:         "set-hostname.service",
			Enabled:      true,
		},
		{
			AssetContent: cloudconfig.EphemeralVarLibDockerMount,
			Name:         "var-lib-docker.mount",
			Enabled:      true,
		},
		// Attach etcd3 dependencies (EBS and ENI).
		{
			AssetContent: cloudconfig.Etcd3AttachDepService,
			Name:         "etcd3-attach-dependencies.service",
			Enabled:      true,
		},
		// Automount etcd EBS volume.
		{
			AssetContent: cloudconfig.AutomountEtcdVolume,
			Name:         "var-lib-etcd.automount",
			Enabled:      true,
		},
		// Mount etcd EBS volume.
		{
			AssetContent: cloudconfig.MountEtcdVolumeAsgMasters,
			Name:         "var-lib-etcd.mount",
			Enabled:      false,
		},
		// Mount log EBS volume.
		{
			AssetContent: cloudconfig.EphemeralVarLogMount,
			Name:         "var-log.mount",
			Enabled:      true,
		},
	}

	var newUnits []k8scloudconfig.UnitAsset

	data := e.templateDataTCCPN()

	for _, fm := range unitsMeta {
		c, err := k8scloudconfig.RenderAssetContent(fm.AssetContent, data)
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

func (e *HAMasterExtension) VerbatimSections() []k8scloudconfig.VerbatimSection {
	newSections := []k8scloudconfig.VerbatimSection{}

	return newSections
}

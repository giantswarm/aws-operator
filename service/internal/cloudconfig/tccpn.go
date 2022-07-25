package cloudconfig

import (
	"context"
	"fmt"
	"net"
	"sync"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v6/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v4/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v14/pkg/template"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys/v3"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/v12/pkg/label"
	"github.com/giantswarm/aws-operator/v12/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/v12/service/controller/key"
	"github.com/giantswarm/aws-operator/v12/service/internal/hamaster"
)

type TCCPNConfig struct {
	Config Config
}

type TCCPN struct {
	config Config
}

func NewTCCPN(config TCCPNConfig) (*TCCPN, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &TCCPN{
		config: config.Config,
	}

	return t, nil
}

func (t *TCCPN) NewPaths(ctx context.Context, obj interface{}) ([]string, error) {
	var err error

	var mappings []hamaster.Mapping
	{
		mappings, err = t.config.HAMaster.Mapping(ctx, obj)
		if hamaster.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "control plane CR")
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var paths []string
	for _, m := range mappings {
		path, err := t.newPath(ctx, obj, m)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		paths = append(paths, path)
	}

	return paths, nil
}

func (t *TCCPN) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	var err error

	var mappings []hamaster.Mapping
	{
		mappings, err = t.config.HAMaster.Mapping(ctx, obj)
		if hamaster.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "control plane CR")
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var templates []string
	for _, mapping := range mappings {
		template, err := t.newTemplate(ctx, obj, mapping)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		templates = append(templates, template)
	}

	return templates, nil
}

func (t *TCCPN) newPath(ctx context.Context, obj interface{}, mapping hamaster.Mapping) (string, error) {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return key.S3ObjectPathTCCPN(&cr, mapping.ID), nil
}

func (t *TCCPN) newTemplate(ctx context.Context, obj interface{}, mapping hamaster.Mapping) (string, error) {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}
	ek, err := t.config.Encrypter.EncryptionKey(ctx, key.ClusterID(&cr))
	if err != nil {
		return "", microerror.Mask(err)
	}
	im, err := t.config.Images.CC(ctx, obj)
	if err != nil {
		return "", microerror.Mask(err)
	}
	v, err := t.config.Images.Versions(ctx, obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var cl infrastructurev1alpha3.AWSCluster
	{
		var list infrastructurev1alpha3.AWSClusterList
		err := t.config.K8sClient.CtrlClient().List(
			ctx,
			&list,
			client.InNamespace(cr.Namespace),
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return "", microerror.Mask(err)
		}

		if len(list.Items) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected 1 CR got %d", len(list.Items))
		}

		cl = list.Items[0]
	}

	var certFiles []certs.File
	var encryptionConfig, serviceAccountV2Pub, serviceAccountV2Priv string
	{
		g := &errgroup.Group{}
		m := sync.Mutex{}

		g.Go(func() error {
			tls, err := t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.APICert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesAPI(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			var err error
			var tls certs.TLS

			switch mapping.ID {
			case 0:
				tls, err = t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.EtcdCert)
			case 1:
				tls, err = t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.Etcd1Cert)
			case 2:
				tls, err = t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.Etcd2Cert)
			case 3:
				tls, err = t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.Etcd3Cert)
			default:
				return microerror.Maskf(executionFailedError, "invalid master id %d", mapping.ID)
			}

			if err != nil {
				return microerror.Mask(err)
			}

			m.Lock()
			certFiles = append(certFiles, certs.NewFilesEtcd(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			tls, err := t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.PrometheusEtcdClientCert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesPrometheusEtcdClient(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			tls, err := t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.ServiceAccountCert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesServiceAccount(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			tls, err := t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.WorkerCert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesWorker(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			var secret v1.Secret
			err := t.config.K8sClient.CtrlClient().Get(
				ctx, client.ObjectKey{
					Name:      key.EncryptionConfigSecretName(key.ClusterID(&cr)),
					Namespace: cr.Namespace,
				},
				&secret)
			if err != nil {
				return microerror.Mask(err)
			}
			encryptionConfig = string(secret.Data[key.EncryptionProviderConfig])

			return nil
		})

		if _, ok := cl.Annotations[annotation.AWSIRSA]; ok {
			// fetch IRSA certs
			g.Go(func() error {
				var secret v1.Secret
				err := t.config.K8sClient.CtrlClient().Get(
					ctx, client.ObjectKey{
						Name:      key.ServiceAccountV2SecretName(key.ClusterID(&cr)),
						Namespace: cr.Namespace,
					},
					&secret)
				if err != nil {
					return microerror.Mask(err)
				}

				serviceAccountV2Pub, err = t.config.Encrypter.Encrypt(ctx, ek, string(secret.Data[key.ServiceAccountV2Pub]))
				if err != nil {
					return microerror.Mask(err)
				}
				serviceAccountV2Priv, err = t.config.Encrypter.Encrypt(ctx, ek, string(secret.Data[key.ServiceAccountV2Priv]))
				if err != nil {
					return microerror.Mask(err)
				}
				return nil
			})
		}

		err := g.Wait()
		if certs.IsTimeout(err) {
			return "", microerror.Maskf(timeoutError, "waited too long for certificates")
		} else if randomkeys.IsTimeout(err) {
			return "", microerror.Maskf(timeoutError, "waited too long for random keys")
		} else if err != nil {
			return "", microerror.Mask(err)
		}
	}

	encryptedEncryptionConfig, err := t.config.Encrypter.Encrypt(ctx, ek, encryptionConfig)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var apiExtraArgs, irsaSAKeyArgs []string
	{
		if key.OIDCClientID(cl) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-client-id=%s", key.OIDCClientID(cl)))
		}
		if key.OIDCIssuerURL(cl) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", key.OIDCIssuerURL(cl)))
		}
		if key.OIDCUsernameClaim(cl) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", key.OIDCUsernameClaim(cl)))
		}
		if key.OIDCGroupsClaim(cl) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", key.OIDCGroupsClaim(cl)))
		}

		// enable IRSA on the api
		if _, ok := cl.Annotations[annotation.AWSIRSA]; ok {
			awsEndpoint := "amazonaws.com"
			if key.IsChinaRegion(key.Region(cl)) {
				awsEndpoint = "amazonaws.com.cn"
			}

			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--service-account-issuer=https://s3.%s.%s/%s-g8s-%s-oidc-pod-identity", key.Region(cl), awsEndpoint, cc.Status.TenantCluster.AWS.AccountID, key.ClusterID(&cr)))
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--api-audiences=sts.%s", awsEndpoint))

			irsaSAKeyArgs = append(irsaSAKeyArgs, "--service-account-key-file=/etc/kubernetes/ssl/service-account-v2-pub.pem")
			irsaSAKeyArgs = append(irsaSAKeyArgs, "--service-account-signing-key-file=/etc/kubernetes/ssl/service-account-v2-priv.pem")

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

	// Pod CIDR should never be nil.
	if key.PodsCIDRBlock(cl) == "" {
		return "", microerror.Maskf(executionFailedError, "Pod CIDR cannot be nil in AWSCluster")
	}

	var controllerManagerExtraArgs []string
	{
		controllerManagerExtraArgs = append(controllerManagerExtraArgs, "--allocate-node-cidrs=true")
		controllerManagerExtraArgs = append(controllerManagerExtraArgs, "--cluster-cidr="+key.PodsCIDRBlock(cl))
		controllerManagerExtraArgs = append(controllerManagerExtraArgs, "--node-cidr-mask-size=25")
	}

	// Allow the actual externalSNAT to be set by the CR.
	var externalSNAT bool
	if key.ExternalSNAT(cl) == nil {
		externalSNAT = t.config.ExternalSNAT
	} else {
		externalSNAT = *key.ExternalSNAT(cl)
	}

	var etcdInitialClusterState string
	{
		if !key.IsAlreadyCreatedCluster(cl) {
			etcdInitialClusterState = k8scloudconfig.InitialClusterStateNew
		} else {
			etcdInitialClusterState = k8scloudconfig.InitialClusterStateExisting
		}
	}

	var multiMasterEnabled bool
	{
		multiMasterEnabled, err = t.config.HAMaster.Enabled(ctx, obj)
		if hamaster.IsNotFound(err) {
			return "", microerror.Maskf(notFoundError, "control plane CR")
		} else if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var params k8scloudconfig.Params
	{
		params = k8scloudconfig.Params{}

		g8sConfig := cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCCPN(&cr, mapping.ID))

		if key.PodsCIDRBlock(cl) != "" {
			_, ipnet, err := net.ParseCIDR(key.PodsCIDRBlock(cl))
			if err != nil {
				return "", microerror.Mask(err)
			}
			g8sConfig.Cluster.Calico.Subnet = ipnet.IP.String()
			_, g8sConfig.Cluster.Calico.CIDR = ipnet.Mask.Size()
		}

		params.BaseDomain = key.TenantClusterBaseDomain(cl)
		params.DisableCalico = true
		params.DisableKubeProxy = false
		params.Cluster = g8sConfig.Cluster
		params.DisableEncryptionAtREST = true
		// Ingress Controller service is not created via ignition.
		// It gets created by the Ingress Controller app if it is installed in the tenant cluster.
		params.DisableIngressControllerService = true
		params.DockerhubToken = t.config.DockerhubToken
		params.EnableAWSCNI = false
		params.EnableCSIMigrationAWS = true
		params.Etcd = k8scloudconfig.Etcd{
			ClientPort:          key.EtcdPort,
			InitialClusterState: etcdInitialClusterState,
			HighAvailability:    multiMasterEnabled,
			NodeName:            key.ControlPlaneEtcdNodeName(mapping.ID),
		}
		// we need to explicitly set InitialCluster for single master, since k8scc qhas different config logic which does nto work for AWS
		if !multiMasterEnabled {
			params.Etcd.InitialCluster = fmt.Sprintf("%s=https://%s.%s:2380", key.ControlPlaneEtcdNodeName(mapping.ID), key.ControlPlaneEtcdNodeName(mapping.ID), key.TenantClusterBaseDomain(cl))
		}
		params.Extension = &TCCPNExtension{
			baseDomain:           key.TenantClusterBaseDomain(cl),
			cc:                   cc,
			cluster:              cl,
			clusterCerts:         certFiles,
			encrypter:            t.config.Encrypter,
			encryptionKey:        ek,
			externalSNAT:         externalSNAT,
			haMasters:            multiMasterEnabled,
			masterID:             mapping.ID,
			encryptionConfig:     encryptedEncryptionConfig,
			registryDomain:       t.config.RegistryDomain,
			serviceAccountv2Priv: serviceAccountV2Priv,
			serviceAccountV2Pub:  serviceAccountV2Pub,
		}
		params.IrsaSAKeyArgs = irsaSAKeyArgs
		params.Kubernetes.Apiserver.CommandExtraArgs = apiExtraArgs
		params.Kubernetes.Kubelet.CommandExtraArgs = kubeletExtraArgs
		params.Kubernetes.ControllerManager.CommandExtraArgs = controllerManagerExtraArgs
		params.ImagePullProgressDeadline = t.config.ImagePullProgressDeadline
		params.RegistryMirrors = t.config.RegistryMirrors
		params.SSOPublicKey = t.config.SSOPublicKey
		params.Images = im
		params.Versions = v

		ignitionPath := k8scloudconfig.GetIgnitionPath(t.config.IgnitionPath)
		params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var templateBody string
	{
		c := k8scloudconfig.CloudConfigConfig{
			Params:   params,
			Template: k8scloudconfig.MasterTemplate,
		}

		cloudConfig, err := k8scloudconfig.NewCloudConfig(c)
		if err != nil {
			return "", microerror.Mask(err)
		}

		err = cloudConfig.ExecuteTemplate()
		if err != nil {
			return "", microerror.Mask(err)
		}

		templateBody = cloudConfig.String()
	}

	return templateBody, nil
}

package cloudconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"

	infrastructurev1alpha3 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha3"
	"github.com/giantswarm/certs/v3/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v11/pkg/template"
	"github.com/giantswarm/k8smetadata/pkg/annotation"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys/v2"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	apiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
)

const IRSAAnnotation = "alpha.aws.giantswarm.io/iam-roles-for-service-accounts"

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
	awsCNIVersion, err := t.config.Images.AWSCNI(ctx, obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var awsCluster infrastructurev1alpha3.AWSCluster
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

		awsCluster = list.Items[0]
	}

	var cluster apiv1alpha3.Cluster
	{
		var list apiv1alpha3.ClusterList
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

		cluster = list.Items[0]
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

	var apiExtraArgs []string
	{
		if key.OIDCClientID(awsCluster) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-client-id=%s", key.OIDCClientID(awsCluster)))
		}
		if key.OIDCIssuerURL(awsCluster) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-issuer-url=%s", key.OIDCIssuerURL(awsCluster)))
		}
		if key.OIDCUsernameClaim(awsCluster) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-username-claim=%s", key.OIDCUsernameClaim(awsCluster)))
		}
		if key.OIDCGroupsClaim(awsCluster) != "" {
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--oidc-groups-claim=%s", key.OIDCGroupsClaim(awsCluster)))
		}

		// enable IRSA on the api
		if _, ok := cluster.Annotations[IRSAAnnotation]; ok {
			apiExtraArgs = append(apiExtraArgs, "--service-account-key-file=/etc/kubernetes/ssl/service-account-key-v2-pub.pem")
			apiExtraArgs = append(apiExtraArgs, "--service-account-signing-key-file=/etc/kubernetes/ssl/service-account-key-v2-priv.pem")
			apiExtraArgs = append(apiExtraArgs, fmt.Sprintf("--service-account-issuer=https://s3-%s.amazonaws.com/%s-%s-oidc-pod-identity", key.Region(awsCluster), cc.Status.TenantCluster.AWS.AccountID, key.ClusterID(&cr)))
			apiExtraArgs = append(apiExtraArgs, "--api-audiences=sts.amazonaws.com")
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

	// Allow the actual externalSNAT to be set by the CR.
	var externalSNAT bool
	if key.ExternalSNAT(awsCluster) == nil {
		externalSNAT = t.config.ExternalSNAT
	} else {
		externalSNAT = *key.ExternalSNAT(awsCluster)
	}

	var etcdInitialClusterState string
	{
		if !key.IsAlreadyCreatedCluster(awsCluster) {
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

	var awsCNIMinimumIPTarget string
	var awsCNIWarmIPTarget string
	{
		awsCNIMinimumIPTarget = key.AWSCNIDefaultMinimumIPTarget
		if v, ok := awsCluster.GetAnnotations()[annotation.AWSCNIMinimumIPTarget]; ok {
			awsCNIMinimumIPTarget = v
		}

		awsCNIWarmIPTarget = key.AWSCNIDefaultWarmIPTarget
		if v, ok := awsCluster.GetAnnotations()[annotation.AWSCNIWarmIPTarget]; ok {
			awsCNIWarmIPTarget = v
		}
	}
	var awsCNIPrefix bool
	{
		if v, ok := awsCluster.GetAnnotations()[annotation.AWSCNIPrefixDelegation]; ok && v == "true" {
			awsCNIPrefix = true
		}
	}
	var awsCNIAdditionalTags string
	{
		var list apiv1alpha3.ClusterList
		err := t.config.K8sClient.CtrlClient().List(
			ctx,
			&list,
			client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
		)
		if err != nil {
			return "", microerror.Mask(err)
		}
		if len(list.Items) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected 1 CR got %d", len(list.Items))
		}

		awsCNIAdditionalTags, err = getCloudTags(list.Items[0].GetLabels())
		if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var params k8scloudconfig.Params
	{
		params = k8scloudconfig.Params{}

		g8sConfig := cmaClusterToG8sConfig(t.config, awsCluster, key.KubeletLabelsTCCPN(&cr, mapping.ID))

		if key.PodsCIDRBlock(awsCluster) != "" {
			_, ipnet, err := net.ParseCIDR(key.PodsCIDRBlock(awsCluster))
			if err != nil {
				return "", microerror.Mask(err)
			}
			g8sConfig.Cluster.Calico.Subnet = ipnet.IP.String()
			_, g8sConfig.Cluster.Calico.CIDR = ipnet.Mask.Size()
		}

		params.BaseDomain = key.TenantClusterBaseDomain(awsCluster)
		params.CalicoPolicyOnly = true
		params.Cluster = g8sConfig.Cluster
		params.DisableEncryptionAtREST = true
		// Ingress Controller service is not created via ignition.
		// It gets created by the Ingress Controller app if it is installed in the tenant cluster.
		params.DisableIngressControllerService = true
		params.DockerhubToken = t.config.DockerhubToken
		params.EnableAWSCNI = true
		params.EnableCSIMigrationAWS = true
		params.Etcd = k8scloudconfig.Etcd{
			ClientPort:          key.EtcdPort,
			InitialClusterState: etcdInitialClusterState,
			HighAvailability:    multiMasterEnabled,
			NodeName:            key.ControlPlaneEtcdNodeName(mapping.ID),
		}
		// we need to explicitly set InitialCluster for single master, since k8scc qhas different config logic which does nto work for AWS
		if !multiMasterEnabled {
			params.Etcd.InitialCluster = fmt.Sprintf("%s=https://%s.%s:2380", key.ControlPlaneEtcdNodeName(mapping.ID), key.ControlPlaneEtcdNodeName(mapping.ID), key.TenantClusterBaseDomain(awsCluster))
		}
		params.Extension = &TCCPNExtension{
			awsCNIAdditionalTags:  awsCNIAdditionalTags,
			awsCNIMinimumIPTarget: awsCNIMinimumIPTarget,
			awsCNIPrefix:          awsCNIPrefix,
			awsCNIVersion:         awsCNIVersion,
			awsCNIWarmIPTarget:    awsCNIWarmIPTarget,
			baseDomain:            key.TenantClusterBaseDomain(awsCluster),
			cc:                    cc,
			cluster:               awsCluster,
			clusterCerts:          certFiles,
			encrypter:             t.config.Encrypter,
			encryptionKey:         ek,
			externalSNAT:          externalSNAT,
			haMasters:             multiMasterEnabled,
			masterID:              mapping.ID,
			encryptionConfig:      encryptedEncryptionConfig,
			registryDomain:        t.config.RegistryDomain,
			serviceAccountv2Priv:  serviceAccountV2Priv,
			serviceAccountV2Pub:   serviceAccountV2Pub,
		}
		params.Kubernetes.Apiserver.CommandExtraArgs = apiExtraArgs
		params.Kubernetes.Kubelet.CommandExtraArgs = kubeletExtraArgs
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

func getCloudTags(labels map[string]string) (string, error) {
	tags := map[string]string{}
	for k, v := range labels {
		if isCloudTagKey(k) {
			tags[trimCloudTagKey(k)] = v
		}
	}

	t, err := json.Marshal(tags)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return string(t), nil
}

// IsCloudTagKey checks if a tag has proper prefix
func isCloudTagKey(tagKey string) bool {
	return strings.HasPrefix(tagKey, key.KeyCloudPrefix)
}

// TrimCloudTagKey trims key cloud prefix from a tag
func trimCloudTagKey(tagKey string) string {
	return strings.TrimPrefix(tagKey, key.KeyCloudPrefix)
}

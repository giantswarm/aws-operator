package cloudconfig

import (
	"context"
	"fmt"
	"net"
	"sync"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
	"github.com/giantswarm/aws-operator/service/internal/hamaster"
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
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// We need to initialize the HA Master state machine. This gives us straight
	// forward access to master ID/AZ mappings. Here the list of master IDs
	// decides if we want to generate certificates for a Tenant Cluster whether a
	// HA Master setup.
	{
		err = t.config.HAMaster.Init(ctx, key.ClusterID(cr))
		if hamaster.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "control plane CR")
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var paths []string
	for !t.config.HAMaster.Reconciled() {
		path, err := t.newPath(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		paths = append(paths, path)
		t.config.HAMaster.Next()
	}

	return paths, nil
}

func (t *TCCPN) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// We need to initialize the HA Master state machine. This gives us straight
	// forward access to master ID/AZ mappings. Here the list of master IDs
	// decides if we want to generate certificates for a Tenant Cluster whether a
	// HA Master setup.
	{
		err = t.config.HAMaster.Init(ctx, key.ClusterID(cr))
		if hamaster.IsNotFound(err) {
			return nil, microerror.Maskf(notFoundError, "control plane CR")
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var templates []string
	for !t.config.HAMaster.Reconciled() {
		template, err := t.newTemplate(ctx, cr)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		templates = append(templates, template)
		t.config.HAMaster.Next()
	}

	return templates, nil
}

func (t *TCCPN) newPath(ctx context.Context, obj interface{}) (string, error) {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}

	return key.S3ObjectPathTCCPN(&cr, t.config.HAMaster.ID()), nil
}

func (t *TCCPN) newTemplate(ctx context.Context, obj interface{}) (string, error) {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return "", microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}
	im, err := t.config.Images.ForRelease(ctx, cr)
	if err != nil {
		return "", microerror.Mask(err)
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
			return "", microerror.Mask(err)
		}

		if len(list.Items) != 1 {
			return "", microerror.Maskf(executionFailedError, "expected 1 CR got %d", len(list.Items))
		}

		cl = list.Items[0]
	}

	var certFiles []certs.File
	var randKeys randomkeys.Cluster
	{
		g := &errgroup.Group{}
		m := sync.Mutex{}

		g.Go(func() error {
			tls, err := t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.APICert)
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

			switch t.config.HAMaster.ID() {
			case 0:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.EtcdCert)
			case 1:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.Etcd1Cert)
			case 2:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.Etcd2Cert)
			case 3:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.Etcd3Cert)
			default:
				return microerror.Maskf(executionFailedError, "invalid master id %d", t.config.HAMaster.ID())
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
			tls, err := t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.ServiceAccountCert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesServiceAccount(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			tls, err := t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.WorkerCert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesWorker(tls)...)
			m.Unlock()

			return nil
		})

		g.Go(func() error {
			k, err := t.config.RandomKeysSearcher.SearchCluster(key.ClusterID(&cr))
			if err != nil {
				return microerror.Mask(err)
			}
			randKeys = k

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

	randomKeyTmplSet, err := renderRandomKeyTmplSet(ctx, t.config.Encrypter, cc.Status.TenantCluster.Encryption.Key, randKeys)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var apiExtraArgs []string
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

		apiExtraArgs = append(apiExtraArgs, t.config.APIExtraArgs...)
	}

	var kubeletExtraArgs []string
	{
		if t.config.PodInfraContainerImage != "" {
			kubeletExtraArgs = append(kubeletExtraArgs, fmt.Sprintf("--pod-infra-container-image=%s", t.config.PodInfraContainerImage))
		}

		kubeletExtraArgs = append(kubeletExtraArgs, t.config.KubeletExtraArgs...)
	}

	var masterSubnet net.IPNet
	for _, az := range cc.Spec.TenantCluster.TCCP.AvailabilityZones {
		if az.Name == t.config.HAMaster.AZ() {
			masterSubnet = az.Subnet.Private.CIDR
			break
		}
	}

	var params k8scloudconfig.Params
	{
		params = k8scloudconfig.DefaultParams()

		g8sConfig := cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCCPN(&cr))
		params.Cluster = g8sConfig.Cluster
		params.DisableEncryptionAtREST = true
		// Ingress Controller service is not created via ignition.
		// It gets created by the Ingress Controller app if it is installed in the tenant cluster.
		params.DisableIngressControllerService = true
		params.EnableAWSCNI = true
		params.EtcdPort = key.EtcdPort
		params.Extension = &TCCPNExtension{
			cc:               cc,
			cluster:          cl,
			clusterCerts:     certFiles,
			encrypter:        t.config.Encrypter,
			encryptionKey:    cc.Status.TenantCluster.Encryption.Key,
			masterSubnet:     masterSubnet,
			masterID:         t.config.HAMaster.ID(),
			randomKeyTmplSet: randomKeyTmplSet,
			registryDomain:   t.config.RegistryDomain,
		}
		params.Hyperkube.Apiserver.Pod.CommandExtraArgs = apiExtraArgs
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = kubeletExtraArgs
		params.ImagePullProgressDeadline = t.config.ImagePullProgressDeadline
		params.RegistryDomain = t.config.RegistryDomain
		params.SSOPublicKey = t.config.SSOPublicKey
		params.Images = im

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

package cloudconfig

import (
	"context"
	"fmt"
	"net"
	"sync"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/hamaster"
	"github.com/giantswarm/aws-operator/service/controller/key"
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

	// We need to determine if we want to generate certificates for a Tenant
	// Cluster with a HA Master setup.
	var haMasterEnabled bool
	{
		haMasterEnabled, err = t.config.HAMaster.Enabled(ctx, key.ClusterID(cr))
		if hamaster.IsNotFound(err) {
			// TODO return package error
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var paths []string
	if haMasterEnabled {
		paths = append(paths, key.S3ObjectPathTCCPN(cr, 1))
		paths = append(paths, key.S3ObjectPathTCCPN(cr, 2))
		paths = append(paths, key.S3ObjectPathTCCPN(cr, 3))
	} else {
		paths = append(paths, key.S3ObjectPathTCCPN(cr, 0))
	}

	return paths, nil
}

func (t *TCCPN) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// We need to determine if we want to generate certificates for a Tenant
	// Cluster with a HA Master setup.
	var haMasterEnabled bool
	{
		haMasterEnabled, err = t.config.HAMaster.Enabled(ctx, key.ClusterID(&cr))
		if hamaster.IsNotFound(err) {
			// TODO return package error
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var templates []string
	if haMasterEnabled {
		t1, err := t.newTemplate(ctx, cr, 1)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		templates = append(templates, t1)

		t2, err := t.newTemplate(ctx, cr, 2)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		templates = append(templates, t2)

		t3, err := t.newTemplate(ctx, cr, 3)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		templates = append(templates, t3)
	} else {
		t0, err := t.newTemplate(ctx, cr, 0)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		templates = append(templates, t0)
	}

	return templates, nil
}

//			CertsSearcher:      config.CertsSearcher,
//			LabelsFunc:         key.KubeletLabelsTCCPN,
//			G8sClient:          config.G8sClient,
//			PathFunc:           key.S3ObjectPathTCCPN,
//			RandomKeysSearcher: config.RandomKeysSearcher,
//			RegistryDomain:     config.RegistryDomain,
func (t *TCCPN) newTemplate(ctx context.Context, cr infrastructurev1alpha2.AWSControlPlane, id int) (string, error) {
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	var cl infrastructurev1alpha2.AWSCluster
	var re *v1alpha1.Release

	var certFiles []certs.File
	var randKeys randomkeys.Cluster
	{
		g := &errgroup.Group{}
		m := sync.Mutex{}

		g.Go(func() error {
			var list infrastructurev1alpha2.AWSClusterList
			err := t.config.K8sClient.CtrlClient().List(
				ctx,
				&list,
				client.InNamespace(cr.Namespace),
				client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
			)
			if err != nil {
				return microerror.Mask(err)
			}

			if len(list.Items) != 1 {
				// TODO return package error
			}

			cl = list.Items[0]

			return nil
		})

		g.Go(func() error {
			rel, err := t.config.K8sClient.G8sClient().ReleaseV1alpha1().Releases().Get(key.ReleaseName(key.ReleaseVersion(&cr)), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			release = rel

			var list releasev1alpha1.ReleaseList
			err := t.config.K8sClient.CtrlClient().List(
				ctx,
				&list,
				client.MatchingLabels{label.Cluster: key.ClusterID(&cr)},
			)
			if err != nil {
				return microerror.Mask(err)
			}

			if len(list.Items) != 1 {
				// TODO return package error
			}

			cl = list.Items[0]

			return nil
		})


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

			switch id {
			case 0:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.EtcdCert)
			case 1:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.Etcd1Cert)
			case 2:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.Etcd2Cert)
			case 3:
				tls, err = t.config.CertsSearcher.SearchTLS(key.ClusterID(&cr), certs.Etcd3Cert)
			default:
				// TODO return package error
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
			// TODO return package error
		} else if randomkeys.IsTimeout(err) {
			// TODO return package error
		} else if err != nil {
			return "", microerror.Mask(err)
		}
	}

	var images k8scloudconfig.Images
	{
		v, err := k8scloudconfig.ExtractComponentVersions(release.Spec.Components)
		if err != nil {
			return "", microerror.Mask(err)
		}

		v.Kubectl = key.KubectlVersion
		v.KubernetesAPIHealthz = key.KubernetesAPIHealthzVersion
		v.KubernetesNetworkSetupDocker = key.K8sSetupNetworkEnvironment
		images = k8scloudconfig.BuildImages(t.config.RegistryDomain, v)
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
	{
		zones := cc.Spec.TenantCluster.TCCP.AvailabilityZones
		for _, az := range zones {
			// TODO is it that a Single Master setup guarantees a single AZ and that a
			// HA Masters setup guarantees 3 AZs?
			if az.Name == key.ControlPlaneAvailabilityZones(cr)[id] {
				masterSubnet = az.Subnet.Private.CIDR
				break
			}
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
		params.EtcdPort = key.EtcdPort
		params.Extension = &TCCPNExtension{
			cc:               cc,
				cluster:        cl,
			clusterCerts:     certFiles,
				encrypter:      t.config.Encrypter,
				encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
				masterSubnet:   masterSubnet,
				masterID:       id,
			randomKeyTmplSet: randomKeyTmplSet,
				registryDomain: t.config.RegistryDomain,
		}
		params.Hyperkube.Apiserver.Pod.CommandExtraArgs = apiExtraArgs
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = kubeletExtraArgs
		params.ImagePullProgressDeadline = t.config.ImagePullProgressDeadline
		params.RegistryDomain = t.config.RegistryDomain
		params.SSOPublicKey = t.config.SSOPublicKey
		params.Images = images

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

package cloudconfig

import (
	"context"
	"fmt"
	"net"
	"sync"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/v3/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/certs/v3/pkg/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v9/pkg/template"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeys/v2"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/giantswarm/aws-operator/pkg/label"
	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/key"
)

type TCNPConfig struct {
	Config Config
}

type TCNP struct {
	config Config
}

func NewTCNP(config TCNPConfig) (*TCNP, error) {
	err := config.Config.Validate()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	t := &TCNP{
		config: config.Config,
	}

	return t, nil
}

func (t *TCNP) NewPaths(ctx context.Context, obj interface{}) ([]string, error) {
	cr, err := meta.Accessor(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return []string{key.S3ObjectPathTCNP(cr)}, nil
}

func (t *TCNP) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	cr, err := key.ToMachineDeployment(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	im, err := t.config.Images.CC(ctx, obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	v, err := t.config.Images.Versions(ctx, obj)
	if err != nil {
		return nil, microerror.Mask(err)
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
			return nil, microerror.Mask(err)
		}

		if len(list.Items) != 1 {
			return nil, microerror.Maskf(executionFailedError, "expected 1 CR got %d", len(list.Items))
		}

		cl = list.Items[0]
	}

	var certFiles []certs.File
	{
		g := &errgroup.Group{}
		m := sync.Mutex{}

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
			tls, err := t.config.CertsSearcher.SearchTLS(ctx, key.ClusterID(&cr), certs.PrometheusEtcdClientCert)
			if err != nil {
				return microerror.Mask(err)
			}
			m.Lock()
			certFiles = append(certFiles, certs.NewFilesPrometheusEtcdClient(tls)...)
			m.Unlock()

			return nil
		})

		err := g.Wait()
		if certs.IsTimeout(err) {
			return nil, microerror.Maskf(timeoutError, "waited too long for certificates")
		} else if randomkeys.IsTimeout(err) {
			return nil, microerror.Maskf(timeoutError, "waited too long for random keys")
		} else if err != nil {
			return nil, microerror.Mask(err)
		}
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
	if key.ExternalSNAT(cl) == nil {
		externalSNAT = t.config.ExternalSNAT
	} else {
		externalSNAT = *key.ExternalSNAT(cl)
	}

	var params k8scloudconfig.Params
	{
		// Default registry, kubernetes, etcd images etcd.
		// Required for proper rending of the templates.
		params = k8scloudconfig.Params{}

		g8sConfig := cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCNP(&cr))

		if key.PodsCIDRBlock(cl) != "" {
			_, ipnet, err := net.ParseCIDR(key.PodsCIDRBlock(cl))
			if err != nil {
				return nil, microerror.Mask(err)
			}
			g8sConfig.Cluster.Calico.Subnet = ipnet.IP.String()
			_, g8sConfig.Cluster.Calico.CIDR = ipnet.Mask.Size()
		}

		params.CalicoPolicyOnly = true
		params.Cluster = g8sConfig.Cluster
		params.DockerhubToken = t.config.DockerhubToken
		params.EnableAWSCNI = true
		params.Extension = &TCNPExtension{
			awsConfigSpec:  cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCNP(&cr)),
			cc:             cc,
			cluster:        cl,
			clusterCerts:   certFiles,
			encrypter:      t.config.Encrypter,
			encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
			externalSNAT:   externalSNAT,
			registryDomain: t.config.RegistryDomain,
		}
		params.Kubernetes.Kubelet.CommandExtraArgs = kubeletExtraArgs
		params.ImagePullProgressDeadline = t.config.ImagePullProgressDeadline
		params.RegistryMirrors = t.config.RegistryMirrors
		params.Images = im
		params.SSOPublicKey = t.config.SSOPublicKey
		params.Versions = v

		ignitionPath := k8scloudconfig.GetIgnitionPath(t.config.IgnitionPath)
		params.Files, err = k8scloudconfig.RenderFiles(ignitionPath, params)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var templateBody string
	{
		c := k8scloudconfig.CloudConfigConfig{
			Params:   params,
			Template: k8scloudconfig.WorkerTemplate,
		}

		cloudConfig, err := k8scloudconfig.NewCloudConfig(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		err = cloudConfig.ExecuteTemplate()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		templateBody = cloudConfig.String()
	}

	return []string{templateBody}, nil
}

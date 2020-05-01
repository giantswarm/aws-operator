package cloudconfig

import (
	"context"
	"fmt"
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
	im, err := t.config.Images.ForRelease(ctx, obj)
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

	var params k8scloudconfig.Params
	{
		// Default registry, kubernetes, etcd images etcd.
		// Required for proper rending of the templates.
		params = k8scloudconfig.DefaultParams()

		params.Cluster = cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCNP(&cr)).Cluster
		params.EnableAWSCNI = true
		params.Extension = &TCNPExtension{
			awsConfigSpec:  cmaClusterToG8sConfig(t.config, cl, key.KubeletLabelsTCNP(&cr)),
			cc:             cc,
			cluster:        cl,
			clusterCerts:   certFiles,
			encrypter:      t.config.Encrypter,
			encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
			registryDomain: t.config.RegistryDomain,
		}
		params.Hyperkube.Kubelet.Docker.CommandExtraArgs = kubeletExtraArgs
		params.Images = im
		params.SSOPublicKey = t.config.SSOPublicKey

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

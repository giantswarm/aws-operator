package cloudconfig

import (
	"context"
	"fmt"
	"net"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	"github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/v_6_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/controller/context/resourcecanceledcontext"
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

//			CertsSearcher:      config.CertsSearcher,
//			LabelsFunc:         key.KubeletLabelsTCCPN,
//			G8sClient:          config.G8sClient,
//			PathFunc:           key.S3ObjectPathTCCPN,
//			RandomKeysSearcher: config.RandomKeysSearcher,
//			RegistryDomain:     config.RegistryDomain,
func (t *TCCPN) NewTemplates(ctx context.Context, obj interface{}) ([]string, error) {
	cr, err := key.ToControlPlane(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cc, err := controllercontext.FromContext(ctx)
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
			// TODO return package error
		}

		cl = list.Items[0]
	}

	releaseVersion := key.ReleaseVersion(cr)

	var cluster infrastructurev1alpha2.AWSCluster
	var clusterCerts certs.Cluster
	var clusterKeys randomkeys.Cluster
	var release *v1alpha1.Release
	{
		g := &errgroup.Group{}

		g.Go(func() error {
			m, err := r.g8sClient.InfrastructureV1alpha2().AWSClusters(cr.GetNamespace()).Get(key.ClusterID(cr), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			cluster = *m

			return nil
		})

		g.Go(func() error {
			certs, err := r.certsSearcher.SearchCluster(key.ClusterID(cr))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterCerts = certs

			return nil
		})

		g.Go(func() error {
			keys, err := r.randomKeysSearcher.SearchCluster(key.ClusterID(cr))
			if err != nil {
				return microerror.Mask(err)
			}
			clusterKeys = keys

			return nil
		})

		g.Go(func() error {
			releaseCR, err := r.g8sClient.ReleaseV1alpha1().Releases().Get(key.ReleaseName(releaseVersion), metav1.GetOptions{})
			if err != nil {
				return microerror.Mask(err)
			}
			release = releaseCR

			return nil
		})

		err = g.Wait()
		if certs.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "certificate secrets are not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if randomkeys.IsTimeout(err) {
			r.logger.LogCtx(ctx, "level", "debug", "message", "random key secrets are not available yet")
			r.logger.LogCtx(ctx, "level", "debug", "message", "canceling resource")
			resourcecanceledcontext.SetCanceled(ctx)
			return nil, nil

		} else if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var images k8scloudconfig.Images
	{
		v, err := k8scloudconfig.ExtractComponentVersions(release.Spec.Components)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		v.Kubectl = key.KubectlVersion
		v.KubernetesAPIHealthz = key.KubernetesAPIHealthzVersion
		v.KubernetesNetworkSetupDocker = key.K8sSetupNetworkEnvironment
		images = k8scloudconfig.BuildImages(r.registryDomain, v)
	}

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
			// TODO is it that a Single Master setup guarantees a single AZ and that a
			// HA Masters setup guarantees 3 AZs?
			if az.Name == key.ControlPlaneAvailabilityZones(cp)[masterID] {
				masterSubnet = az.Subnet.Private.CIDR
				break
			}
		}
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
		params.Extension = &TCCPNExtension{
			baseExtension: baseExtension{
				cluster:        cr,
				encrypter:      t.config.Encrypter,
				encryptionKey:  cc.Status.TenantCluster.Encryption.Key,
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

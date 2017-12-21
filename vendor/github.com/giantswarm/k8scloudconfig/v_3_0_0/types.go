package v_3_0_0

import "github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"

const (
	defaultHyperkubeApiserverBindAddress = "${DEFAULT_IPV4}"
)

type Params struct {
	Cluster v1alpha1.Cluster
	// DisableCalico flag when set removes all calico related Kubernetes
	// manifests from the cloud config together with their initialization.
	DisableCalico bool
	// Hyperkube allows to pass extra `docker run` and `command` arguments
	// to hyperkube image commands. This allows to e.g. add cloud provider
	// extensions.
	Hyperkube Hyperkube
	// EtcdPort allows the Etcd port to be specified.
	// aws-operator sets this to the Etcd listening port so Calico on the
	// worker nodes can access via a CNAME record to the master.
	EtcdPort  int
	Extension Extension
	// ExtraManifests allows to specify extra Kubernetes manifests in
	// /opt/k8s-addons script. The manifests are applied after calico is
	// ready.
	//
	// The general use-case is to create a manifest file with Extension and
	// then apply the manifest by adding it to ExtraManifests.
	ExtraManifests []string
	// MasterAPIDomain is a value of domain passed to various Kubernetes
	// services. When MasterAPIDomain is empty value of
	// Cluster.Kubernetes.API.Domain is passed.
	//
	// NOTE This is a work around limitation of Azure load balancers.
	// Hopefully Load Balancer Standard SKU will allow to get rid of that.
	//
	// azure-operator sets that to 127.0.0.1. Other operators leave it empty.
	MasterAPIDomain string
	Node            v1alpha1.ClusterNode
}

type Hyperkube struct {
	Apiserver         HyperkubeApiserver
	ControllerManager HyperkubeControllerManager
	Kubelet           HyperkubeKubelet
}

type HyperkubeApiserver struct {
	// BindAddress is a value of the --bind-address flag passed to the
	// hyperkube apiserver. When BindAddress is empty value of
	// `${DEFAULT_IPV4}` will be passed.
	//
	// NOTE This is a work around limitation of Azure load balancers.
	// Hopefully Load Balancer Standard SKU will allow to get rid of that.
	//
	// azure-operator sets that to 0.0.0.0. Other operators leave it empty.
	BindAddress string
	Docker      HyperkubeDocker
}

type HyperkubeControllerManager struct {
	Docker HyperkubeDocker
}

type HyperkubeKubelet struct {
	Docker HyperkubeDocker
}

type HyperkubeDocker struct {
	RunExtraArgs     []string
	CommandExtraArgs []string
}

type FileMetadata struct {
	AssetContent string
	Path         string
	Owner        string
	Encoding     string
	Permissions  int
}

type FileAsset struct {
	Metadata FileMetadata
	Content  []string
}

type UnitMetadata struct {
	AssetContent string
	Name         string
	Enable       bool
	Command      string
}

type UnitAsset struct {
	Metadata UnitMetadata
	Content  []string
}

// VerbatimSection is a blob of YAML we want to add to the
// CloudConfig, with no variable interpolation.
type VerbatimSection struct {
	Name    string
	Content string
}

type Extension interface {
	Files() ([]FileAsset, error)
	Units() ([]UnitAsset, error)
	VerbatimSections() []VerbatimSection
}

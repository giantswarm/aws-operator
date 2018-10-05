package v_3_2_4

import (
	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

type Params struct {
	// APIServerEncryptionKey is AES-CBC with PKCS#7 padding key to encrypt API
	// etcd data.
	APIServerEncryptionKey string
	Cluster                v1alpha1.Cluster
	// DisableCalico flag when set removes all calico related Kubernetes
	// manifests from the cloud config together with their initialization.
	DisableCalico bool
	// DisableEncryptionAtREST flag when set removes all manifests from the cloud
	// config related to Kubernetes encryption at REST.
	DisableEncryptionAtREST bool
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
	Node           v1alpha1.ClusterNode
}

func (p *Params) Validate() error {
	return nil
}

type Hyperkube struct {
	Apiserver         HyperkubeApiserver
	ControllerManager HyperkubeControllerManager
	Kubelet           HyperkubeKubelet
}

type HyperkubeApiserver struct {
	Pod HyperkubePod
}

type HyperkubeControllerManager struct {
	Pod HyperkubePod
}

type HyperkubeKubelet struct {
	Docker HyperkubeDocker
}

type HyperkubeDocker struct {
	RunExtraArgs     []string
	CommandExtraArgs []string
}

type HyperkubePod struct {
	HyperkubePodHostExtraMounts []HyperkubePodHostMount
	CommandExtraArgs            []string
}

type HyperkubePodHostMount struct {
	Name     string
	Path     string
	ReadOnly bool
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

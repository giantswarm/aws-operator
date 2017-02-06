package clusterspec

type ClusterSpec struct {
	Customer  string `json:"customer"`
	ClusterId string `json:"clusterId"`

	K8sVersion               string `json:"k8sVersion"`
	K8sVmVersion             string `json:"k8sVmVersion"`
	CertctlVersion           string `json:"certctlVersion"`
	PingVersion              string `json:"pingVersion"`
	IngressControllerVersion string `json:"ingressControllerVersion"`
	// Kubectl is used in the master/worker to allow dynamic updates creating configmaps
	KubectlVersion string `json:"kubectlVersion"`

	FlannelConfiguration FlannelConfiguration `json::"flannelConfiguration"`

	Certificates Certificates `json:"certificates"`

	Worker Worker `json:"worker"`
	Master Master `json:"master"`

	IngressController IngressController `json:"ingressController"`

	GiantnetesConfiguration GiantnetesConfiguration `json:"giantnetesConfiguration"`
}

type GiantnetesConfiguration struct {
	EtcdPort         string `json:"etcdPort"`
	NetworkInterface string `json:"networkInterface"`
	HostSubnetRange  string `json:"hostSubnetRange"`
	DnsIp            string `json:"dnsIp"`
	ApiIp            string `json:"apiIp"`
	Domain           string `json:"domain"`
	VaultAddr        string `json:"vaultAddr"`
	CloudflareDomain string `json:"cloudflareDomain"`
}

type FlannelConfiguration struct {
	Version        string `json:"version"`
	ClusterVni     int32  `json:"clusterVni,omitempty"`
	ClusterBackend string `json:"clusterBackend"`
	ClusterNetwork string `json:"clusterNetwork"`
}

type IngressController struct {
	KempVsIp              string `json:"kempVsIp"`
	KempVsName            string `json:"kempVsName"`
	KempVsPorts           string `json:"kempVsPorts"`
	KempVsSslAcceleration string `json:"kempVsSslAcceleration"`
	KempRsPort            string `json:"kempRsPort"`
	KempUser              string `json:"kempUser"`
	KempEndpoint          string `json:"kempEndpoint"`
	KempPassword          string `json:"kempPassword"`
	KempVsCheckPort       string `json:"kempVsCheckPort"`
	CloudflareIp          string `json:"cloudflareIp"`
	CloudflareDomain      string `json:"cloudflareDomain"`
	CloudflareToken       string `json:"cloudflareToken"`
	CloudflareEmail       string `json:"cloudflareEmail"`
}

type Certificates struct {
	VaultToken        string `json:"vaultToken"`
	ApiAltNames       string `json:"apiAltNames"`
	MasterServiceName string `json:"masterServiceName"`
}

type Machine struct {
	Registry            string       `json:"registry"`
	Capabilities        Capabilities `json:"capabilities"`
	NetworkSetupVersion string       `json:"networkSetupVersion"`
	DockerExtraArgs     string       `json:"dockerExtraArgs,omitempty"`
}

type Capabilities struct {
	Memory   string `json:"memory"`
	CpuCores int32  `json:"cpuCores"`
}

type Worker struct {
	Machine
	Replicas int32 `json:"replicas,omitempty"`

	K8sCalicoMtu      string `json:"k8sCalicoMtu"`
	NodeLabels        string `json:"nodeLabels,omitempty"`
	WorkerServicePort string `json:"workerServicePort"`

	MasterPort string `json:"masterPort"`
	Kubernetes
}

type Master struct {
	Machine

	CalicoSubnet string `json:"calicoSubnet"`
	CalicoCidr   string `json:"calicoCidr"`

	ClusterIpRange  string `json:"clusterIpRange"`
	ClusterIpSubnet string `json:"clusterIpSubnet"`
	Kubernetes
}

type Kubernetes struct {
	Domain           string `json:"domain"`
	EtcdDomainName   string `json:"etcdDomainName"`
	MasterDomainName string `json:"masterDomainName"`
	DnsIp            string `json:"dnsIp"`
	InsecurePort     string `json:"insecurePort"`
	SecurePort       string `json:"securePort"`
}

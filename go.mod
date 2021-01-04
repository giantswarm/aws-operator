module github.com/giantswarm/aws-operator

go 1.15

require (
	github.com/aws/amazon-vpc-cni-k8s v1.7.8
	github.com/aws/aws-sdk-go v1.36.19
	github.com/dylanmei/iso8601 v0.1.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions/v3 v3.13.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/badnodedetector v1.0.1
	github.com/giantswarm/certs/v3 v3.1.0
	github.com/giantswarm/ipam v0.2.0
	github.com/giantswarm/k8sclient/v5 v5.0.0
	github.com/giantswarm/k8scloudconfig/v9 v9.3.0
	github.com/giantswarm/kubelock/v2 v2.0.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.3.0
	github.com/giantswarm/microkit v0.2.2
	github.com/giantswarm/micrologger v0.4.0
	github.com/giantswarm/operatorkit/v4 v4.0.0
	github.com/giantswarm/randomkeys/v2 v2.0.0
	github.com/giantswarm/tenantcluster/v4 v4.0.0
	github.com/giantswarm/to v0.3.0
	github.com/google/go-cmp v0.5.4
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.9.0
	github.com/spf13/afero v1.3.1 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	k8s.io/api v0.18.9
	k8s.io/apiextensions-apiserver v0.18.9
	k8s.io/apimachinery v0.18.9
	k8s.io/client-go v0.18.9
	sigs.k8s.io/cluster-api v0.3.11
	sigs.k8s.io/controller-runtime v0.6.4
)

replace sigs.k8s.io/cluster-api v0.3.11 => github.com/giantswarm/cluster-api v0.3.10-gs

module github.com/giantswarm/aws-operator

go 1.14

require (
	github.com/aws/amazon-vpc-cni-k8s v1.6.3
	github.com/aws/aws-sdk-go v1.33.20
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions v0.4.20
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/certs/v2 v2.0.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/ipam v0.2.0
	github.com/giantswarm/k8sclient/v3 v3.1.1
	github.com/giantswarm/k8scloudconfig/v7 v7.0.5
	github.com/giantswarm/kubelock v0.2.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.1
	github.com/giantswarm/microkit v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit v1.2.0
	github.com/giantswarm/randomkeys v0.2.0
	github.com/giantswarm/tenantcluster/v2 v2.0.0
	github.com/giantswarm/to v0.3.0
	github.com/google/go-cmp v0.5.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/afero v1.3.1 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	k8s.io/api v0.17.8
	k8s.io/apiextensions-apiserver v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	sigs.k8s.io/cluster-api v0.3.8
	sigs.k8s.io/controller-runtime v0.5.9
)

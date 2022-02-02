module github.com/giantswarm/aws-operator

go 1.15

require (
	github.com/aws/amazon-vpc-cni-k8s v1.10.1
	github.com/aws/aws-sdk-go v1.42.38
	github.com/dylanmei/iso8601 v0.1.0
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions/v3 v3.40.0
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/badnodedetector v1.0.1
	github.com/giantswarm/certs/v3 v3.1.1
	github.com/giantswarm/ipam v0.3.0
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/k8scloudconfig/v11 v11.0.1
	github.com/giantswarm/k8smetadata v0.9.2
	github.com/giantswarm/kubelock/v3 v3.0.0
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/operatorkit/v5 v5.0.0
	github.com/giantswarm/randomkeys/v2 v2.1.0
	github.com/giantswarm/tenantcluster/v4 v4.1.0
	github.com/giantswarm/to v0.4.0
	github.com/google/go-cmp v0.5.7
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.12.0
	github.com/spf13/viper v1.10.1
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	k8s.io/api v0.18.19
	k8s.io/apiextensions-apiserver v0.18.19
	k8s.io/apimachinery v0.18.19
	k8s.io/client-go v0.18.19
	sigs.k8s.io/cluster-api v0.4.1
	sigs.k8s.io/controller-runtime v0.6.4
)

replace (
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)

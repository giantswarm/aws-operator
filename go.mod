module github.com/giantswarm/aws-operator

go 1.16

require (
	github.com/aws/amazon-vpc-cni-k8s v1.10.2
	github.com/aws/aws-sdk-go v1.43.1
	github.com/dylanmei/iso8601 v0.1.0
	github.com/getsentry/sentry-go v0.11.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions/v3 v3.39.0
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/badnodedetector v1.0.1
	github.com/giantswarm/certs/v3 v3.1.1
	github.com/giantswarm/ipam v0.3.0
	github.com/giantswarm/k8sclient/v5 v5.12.0
	github.com/giantswarm/k8scloudconfig/v11 v11.1.2
	github.com/giantswarm/k8smetadata v0.10.1
	github.com/giantswarm/kubelock/v3 v3.0.0
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v0.6.0
	github.com/giantswarm/operatorkit/v5 v5.0.0
	github.com/giantswarm/randomkeys/v2 v2.1.0
	github.com/giantswarm/tenantcluster/v4 v4.1.0
	github.com/giantswarm/to v0.4.0
	github.com/gobuffalo/flect v0.2.3 // indirect
	github.com/google/go-cmp v0.5.7
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.16.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.12.1
	github.com/spf13/viper v1.10.1
	golang.org/x/mod v0.5.1 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b // indirect
	sigs.k8s.io/cluster-api v1.0.4
	sigs.k8s.io/controller-runtime v0.10.3
)

replace (
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	sigs.k8s.io/cluster-api => github.com/giantswarm/cluster-api v0.3.13-gs
)

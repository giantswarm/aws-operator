module github.com/giantswarm/aws-operator/v14

go 1.18

require (
	github.com/aws/amazon-vpc-cni-k8s v1.12.1
	github.com/aws/aws-sdk-go v1.44.189
	github.com/blang/semver v3.5.1+incompatible
	github.com/dylanmei/iso8601 v0.1.0
	github.com/ghodss/yaml v1.0.1-0.20220118164431-d8423dcdf344
	github.com/giantswarm/apiextensions/v6 v6.4.1
	github.com/giantswarm/backoff v1.0.0
	github.com/giantswarm/badnodedetector/v2 v2.0.0
	github.com/giantswarm/certs/v4 v4.0.0
	github.com/giantswarm/ipam v0.3.0
	github.com/giantswarm/k8sclient/v7 v7.0.1
	github.com/giantswarm/k8scloudconfig/v15 v15.5.0
	github.com/giantswarm/k8smetadata v0.19.0
	github.com/giantswarm/kubelock/v4 v4.0.0
	github.com/giantswarm/microendpoint v1.0.0
	github.com/giantswarm/microerror v0.4.0
	github.com/giantswarm/microkit v1.0.0
	github.com/giantswarm/micrologger v1.0.0
	github.com/giantswarm/operatorkit/v7 v7.1.0
	github.com/giantswarm/randomkeys/v3 v3.0.0
	github.com/giantswarm/release-operator/v3 v3.2.0
	github.com/giantswarm/tenantcluster/v6 v6.0.0
	github.com/giantswarm/to v0.4.0
	github.com/google/go-cmp v0.5.9
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.14.0
	github.com/spf13/viper v1.15.0
	golang.org/x/sync v0.1.0
	k8s.io/api v0.24.1
	k8s.io/apiextensions-apiserver v0.24.1
	k8s.io/apimachinery v0.24.1
	k8s.io/client-go v0.24.1
	sigs.k8s.io/cluster-api v1.1.4
	sigs.k8s.io/controller-runtime v0.12.1
)

require (
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.2 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/getsentry/sentry-go v0.12.0 // indirect
	github.com/giantswarm/exporterkit v1.0.0 // indirect
	github.com/giantswarm/microstorage v0.2.0 // indirect
	github.com/giantswarm/versionbundle v1.0.0 // indirect
	github.com/go-kit/kit v0.12.0 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gobuffalo/flect v0.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.6.9 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/cobra v1.4.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/term v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/time v0.1.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/resty.v1 v1.12.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/component-base v0.24.1 // indirect
	k8s.io/klog/v2 v2.60.1 // indirect
	k8s.io/kube-openapi v0.0.0-20220603121420-31174f50af60 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20220525155127-227cbc7cc124 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.6.16
	github.com/containernetworking/cni => github.com/containernetworking/cni v1.1.2
	github.com/coreos/etcd v3.3.13+incompatible => github.com/coreos/etcd v3.3.25+incompatible
	github.com/deislabs/oras => github.com/deislabs/oras v0.16.0
	github.com/dgrijalva/jwt-go => github.com/dgrijalva/jwt-go/v4 v4.0.0-preview1
	github.com/getsentry/sentry-go => github.com/getsentry/sentry-go v0.17.0
	github.com/gogo/protobuf v1.3.1 => github.com/gogo/protobuf v1.3.2
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.14.4
	github.com/labstack/echo/v4 => github.com/labstack/echo/v4 v4.10.0
	github.com/microcosm-cc/bluemonday => github.com/microcosm-cc/bluemonday v1.0.22
	github.com/nats-io/nats-server/v2 => github.com/nats-io/nats-server/v2 v2.9.11
	github.com/opencontainers/image-spec => github.com/opencontainers/image-spec v1.0.2
	github.com/opencontainers/runc => github.com/opencontainers/runc v1.1.4
	github.com/pkg/sftp => github.com/pkg/sftp v1.13.5
	github.com/valyala/fasthttp v1.6.0 => github.com/valyala/fasthttp v1.38.0
	helm.sh/helm/v3 => helm.sh/helm/v3 v3.11.0
	sigs.k8s.io/cluster-api => sigs.k8s.io/cluster-api v1.0.5
)

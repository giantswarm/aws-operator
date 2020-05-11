module github.com/giantswarm/aws-operator

go 1.13

require (
	github.com/aws/amazon-vpc-cni-k8s v1.6.0
	github.com/aws/aws-sdk-go v1.29.32
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/giantswarm/apiextensions v0.3.6
	github.com/giantswarm/apprclient v0.2.0
	github.com/giantswarm/backoff v0.2.0
	github.com/giantswarm/certs v0.2.0
	github.com/giantswarm/e2e-harness v0.2.0
	github.com/giantswarm/e2eclients v0.2.0
	github.com/giantswarm/e2esetup v0.1.0
	github.com/giantswarm/e2etemplates v0.2.0
	github.com/giantswarm/e2etests v0.1.0
	github.com/giantswarm/errors v0.2.3
	github.com/giantswarm/exporterkit v0.2.0
	github.com/giantswarm/helmclient v0.2.2
	github.com/giantswarm/ipam v0.2.0
	github.com/giantswarm/k8sclient v0.2.0
	github.com/giantswarm/k8scloudconfig/v6 v6.1.1
	github.com/giantswarm/kubelock v0.2.0
	github.com/giantswarm/microendpoint v0.2.0
	github.com/giantswarm/microerror v0.2.0
	github.com/giantswarm/microkit v0.2.1
	github.com/giantswarm/micrologger v0.3.1
	github.com/giantswarm/operatorkit v0.2.0
	github.com/giantswarm/randomkeys v0.2.0
	github.com/giantswarm/tenantcluster v0.2.0
	github.com/giantswarm/to v0.2.0
	github.com/giantswarm/valuemodifier v0.2.0 // indirect
	github.com/giantswarm/versionbundle v0.2.0
	github.com/google/go-cmp v0.4.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.5.1
	github.com/spf13/afero v1.2.2
	github.com/spf13/viper v1.7.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	k8s.io/api v0.17.2
	k8s.io/apiextensions-apiserver v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	sigs.k8s.io/cluster-api v0.3.2
	sigs.k8s.io/controller-runtime v0.5.2
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.5-beta.1
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191114103151-9ca1dc586682
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191114110141-0a35778df828
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191114112024-4bbba8331835
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191114111741-81bb9acf592d
	k8s.io/code-generator => k8s.io/code-generator v0.16.5-beta.1
	k8s.io/component-base => k8s.io/component-base v0.0.0-20191114102325-35a9586014f7
	k8s.io/cri-api => k8s.io/cri-api v0.16.5-beta.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191114112310-0da609c4ca2d
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191114103820-f023614fb9ea
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191114111510-6d1ed697a64b
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191114110717-50a77e50d7d9
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191114111229-2e90afcb56c7
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191114113550-6123e1c827f7
	k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191114110954-d67a8e7e2200
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191114112655-db9be3e678bb
	k8s.io/metrics => k8s.io/metrics v0.0.0-20191114105837-a4a2842dc51b
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191114104439-68caf20693ac
)

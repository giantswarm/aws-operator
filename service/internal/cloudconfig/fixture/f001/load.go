package f001

import (
	"io/ioutil"
	"path/filepath"

	infrastructurev1alpha2 "github.com/giantswarm/apiextensions/pkg/apis/infrastructure/v1alpha2"
	releasev1alpha1 "github.com/giantswarm/apiextensions/pkg/apis/release/v1alpha1"
	"github.com/giantswarm/certs/v2/pkg/certs"
	"github.com/giantswarm/microerror"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	apiv1alpha2 "sigs.k8s.io/cluster-api/api/v1alpha2"
)

const (
	id = "8y5ck"
)

func MustLoadTLS() map[string]map[certs.Cert]certs.TLS {
	var err error

	var scheme *runtime.Scheme
	{
		scheme = runtime.NewScheme()

		err = apiv1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = corev1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = infrastructurev1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = releasev1alpha1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
	}

	var decoder runtime.Decoder
	{
		codec := serializer.NewCodecFactory(scheme)
		decoder = codec.UniversalDeserializer()
	}

	tls := map[string]map[certs.Cert]certs.TLS{
		id: map[certs.Cert]certs.TLS{
			certs.APICert:                certs.TLS{},
			certs.AppOperatorAPICert:     certs.TLS{},
			certs.AWSOperatorAPICert:     certs.TLS{},
			certs.CalicoEtcdClientCert:   certs.TLS{},
			certs.ClusterOperatorAPICert: certs.TLS{},
			certs.Etcd1Cert:              certs.TLS{},
			certs.Etcd2Cert:              certs.TLS{},
			certs.Etcd3Cert:              certs.TLS{},
			certs.NodeOperatorCert:       certs.TLS{},
			certs.PrometheusCert:         certs.TLS{},
			certs.ServiceAccountCert:     certs.TLS{},
			certs.WorkerCert:             certs.TLS{},
		},
	}

	for i, clusterCerts := range tls {
		for c, t := range clusterCerts {
			s := &corev1.Secret{}

			_, _, err := decoder.Decode(mustLoad(c.String()), nil, s)
			if err != nil {
				panic(err)
			}

			err = fillTLSFromSecret(&t, s, i, c)
			if err != nil {
				panic(err)
			}
			tls[i][c] = t
		}
	}

	return tls
}

func MustLoadRandomKey() *corev1.Secret {
	var err error

	var scheme *runtime.Scheme
	{
		scheme = runtime.NewScheme()

		err = apiv1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = corev1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = infrastructurev1alpha2.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		err = releasev1alpha1.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
	}

	var decoder runtime.Decoder
	{
		codec := serializer.NewCodecFactory(scheme)
		decoder = codec.UniversalDeserializer()
	}

	sec := &corev1.Secret{}
	{
		_, _, err := decoder.Decode(mustLoad("encryption"), nil, sec)
		if err != nil {
			panic(err)
		}
	}

	return sec
}

const (
	// certificateLabel is the label used in the secret to identify a secret
	// containing the certificate.
	certificateLabel = "giantswarm.io/certificate"
	// clusterLabel is the label used in the secret to identify a secret
	// containing the certificate.
	clusterLabel = "giantswarm.io/cluster"
)

func fillTLSFromSecret(tls *certs.TLS, secret *corev1.Secret, cluster string, cert certs.Cert) error {
	{
		var l string

		l = secret.Labels[clusterLabel]
		if cluster != l {
			return microerror.Maskf(invalidSecretError, "expected cluster = %q, got %q", cluster, l)
		}
		l = secret.Labels[certificateLabel]
		if string(cert) != l {
			return microerror.Maskf(invalidSecretError, "expected certificate = %q, got %q", cert, l)
		}
	}

	{
		var ok bool

		if tls.CA, ok = secret.Data["ca"]; !ok {
			return microerror.Maskf(invalidSecretError, "%q key missing", "ca")
		}
		if tls.Crt, ok = secret.Data["crt"]; !ok {
			return microerror.Maskf(invalidSecretError, "%q key missing", "crt")
		}
		if tls.Key, ok = secret.Data["key"]; !ok {
			return microerror.Maskf(invalidSecretError, "%q key missing", "key")
		}
	}

	return nil
}

func mustLoad(f string) []byte {
	p := filepath.Join("/Users/xh3b4sd/go/src/github.com/giantswarm/aws-operator/service/internal/cloudconfig/fixture/f001", f)
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	return b
}

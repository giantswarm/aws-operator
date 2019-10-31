package unittest

import (
	"github.com/giantswarm/certs"
)

func DefaultCerts() certs.Cluster {
	return certs.Cluster{
		APIServer: certs.TLS{
			CA:  []byte("api-server-ca"),
			Crt: []byte("api-server-crt"),
			Key: []byte("api-server-key"),
		},
		CalicoEtcdClient: certs.TLS{
			CA:  []byte("calico-etcd-client-ca"),
			Crt: []byte("calico-etcd-client-crt"),
			Key: []byte("calico-etcd-client-key"),
		},
		EtcdServer: certs.TLS{
			CA:  []byte("etcd-server-ca"),
			Crt: []byte("etcd-server-crt"),
			Key: []byte("etcd-server-key"),
		},
		ServiceAccount: certs.TLS{
			CA:  []byte("service-account-ca"),
			Crt: []byte("service-account-crt"),
			Key: []byte("service-account-key"),
		},
		Worker: certs.TLS{
			CA:  []byte("worker-ca"),
			Crt: []byte("worker-crt"),
			Key: []byte("worker-key"),
		},
	}
}

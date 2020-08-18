package unittest

import (
	"github.com/giantswarm/certs/v3/pkg/certs"
)

func DefaultCerts() []certs.File {
	var list []certs.File

	list = append(list, certs.NewFilesAPI(certs.TLS{
		CA:  []byte("api-server-ca"),
		Crt: []byte("api-server-crt"),
		Key: []byte("api-server-key"),
	})...)

	list = append(list, certs.NewFilesCalicoEtcdClient(certs.TLS{
		CA:  []byte("api-server-ca"),
		Crt: []byte("api-server-crt"),
		Key: []byte("api-server-key"),
	})...)

	list = append(list, certs.NewFilesEtcd(certs.TLS{
		CA:  []byte("api-server-ca"),
		Crt: []byte("api-server-crt"),
		Key: []byte("api-server-key"),
	})...)

	list = append(list, certs.NewFilesServiceAccount(certs.TLS{
		CA:  []byte("api-server-ca"),
		Crt: []byte("api-server-crt"),
		Key: []byte("api-server-key"),
	})...)

	list = append(list, certs.NewFilesWorker(certs.TLS{
		CA:  []byte("api-server-ca"),
		Crt: []byte("api-server-crt"),
		Key: []byte("api-server-key"),
	})...)

	return list
}

package create

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/k8scloudconfig"
	microerror "github.com/giantswarm/microkit/error"
)

// PEM encoded TLS assets.
type rawTLSAssets struct {
	APIServerCACrt    []byte
	APIServerKey      []byte
	APIServerCrt      []byte
	CalicoClientCACrt []byte
	CalicoClientKey   []byte
	CalicoClientCrt   []byte
	EtcdServerCACrt   []byte
	EtcdServerKey     []byte
	EtcdServerCrt     []byte
}

// Encrypted PEM encoded TLS assets
type encryptedTLSAssets struct {
	APIServerCACrt    []byte
	APIServerKey      []byte
	APIServerCrt      []byte
	CalicoClientCACrt []byte
	CalicoClientKey   []byte
	CalicoClientCrt   []byte
	EtcdServerCACrt   []byte
	EtcdServerKey     []byte
	EtcdServerCrt     []byte
}

func readRawTLSAssets(tlsAssetsDir string) (*rawTLSAssets, error) {
	r := new(rawTLSAssets)
	files := []struct {
		name          string
		ca, key, cert *[]byte
	}{
		{"apiserver", &r.APIServerCACrt, &r.APIServerKey, &r.APIServerCrt},
		{"calico/client", &r.CalicoClientCACrt, &r.CalicoClientKey, &r.CalicoClientCrt},
		{"etcd/server", &r.EtcdServerCACrt, &r.EtcdServerKey, &r.EtcdServerCrt},
	}
	for _, file := range files {
		caPath := filepath.Join(tlsAssetsDir, file.name+"-ca.pem")
		keyPath := filepath.Join(tlsAssetsDir, file.name+"-key.pem")
		certPath := filepath.Join(tlsAssetsDir, file.name+"-crt.pem")

		caData, err := ioutil.ReadFile(caPath)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		*file.ca = caData

		certData, err := ioutil.ReadFile(certPath)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		*file.cert = certData

		keyData, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, microerror.MaskAny(err)
		}
		*file.key = keyData
	}
	return r, nil
}

func (r *rawTLSAssets) encrypt(svc *kms.KMS, kmsKeyARN string) (*encryptedTLSAssets, error) {
	var err error
	encrypt := func(data []byte) []byte {
		if err != nil {
			return []byte{}
		}

		encryptInput := kms.EncryptInput{
			KeyId:     aws.String(kmsKeyARN),
			Plaintext: data,
		}

		var encryptOutput *kms.EncryptOutput
		if encryptOutput, err = svc.Encrypt(&encryptInput); err != nil {
			return []byte{}
		}
		return encryptOutput.CiphertextBlob
	}
	encryptedAssets := encryptedTLSAssets{
		APIServerCACrt:    encrypt(r.APIServerCACrt),
		APIServerKey:      encrypt(r.APIServerKey),
		APIServerCrt:      encrypt(r.APIServerCrt),
		CalicoClientCACrt: encrypt(r.CalicoClientCACrt),
		CalicoClientKey:   encrypt(r.CalicoClientKey),
		CalicoClientCrt:   encrypt(r.CalicoClientCrt),
		EtcdServerCACrt:   encrypt(r.EtcdServerCACrt),
		EtcdServerKey:     encrypt(r.EtcdServerKey),
		EtcdServerCrt:     encrypt(r.EtcdServerCrt),
	}
	if err != nil {
		return nil, microerror.MaskAny(err)
	}
	return &encryptedAssets, nil
}

func (r *encryptedTLSAssets) compact() (*cloudconfig.CompactTLSAssets, error) {
	var err error
	compact := func(data []byte) string {
		if err != nil {
			return ""
		}

		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		if _, err := gzw.Write(data); err != nil {
			return ""
		}
		if err := gzw.Close(); err != nil {
			return ""
		}
		return base64.StdEncoding.EncodeToString(buf.Bytes())
	}

	compactAssets := cloudconfig.CompactTLSAssets{
		APIServerCACrt:    compact(r.APIServerCACrt),
		APIServerKey:      compact(r.APIServerKey),
		APIServerCrt:      compact(r.APIServerCrt),
		CalicoClientCACrt: compact(r.CalicoClientCACrt),
		CalicoClientKey:   compact(r.CalicoClientKey),
		CalicoClientCrt:   compact(r.CalicoClientCrt),
		EtcdServerCACrt:   compact(r.EtcdServerCACrt),
		EtcdServerKey:     compact(r.EtcdServerKey),
		EtcdServerCrt:     compact(r.EtcdServerCrt),
	}

	if err != nil {
		return nil, microerror.MaskAny(err)
	}
	return &compactAssets, nil
}

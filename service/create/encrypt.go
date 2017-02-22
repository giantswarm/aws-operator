package create

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	microerror "github.com/giantswarm/microkit/error"
)

// PEM encoded TLS assets.
type rawTLSAssets struct {
	APIServerCACert    []byte
	APIServerKey       []byte
	APIServerCert      []byte
	CalicoClientCACert []byte
	CalicoClientKey    []byte
	CalicoClientCert   []byte
	EtcdServerCACert   []byte
	EtcdServerKey      []byte
	EtcdServerCert     []byte
}

// Encrypted PEM encoded TLS assets
type encryptedTLSAssets struct {
	APIServerCACert    []byte
	APIServerKey       []byte
	APIServerCert      []byte
	CalicoClientCACert []byte
	CalicoClientKey    []byte
	CalicoClientCert   []byte
	EtcdServerCACert   []byte
	EtcdServerKey      []byte
	EtcdServerCert     []byte
}

// PEM -> encrypted -> gzip -> base64 encoded TLS assets.
type CompactTLSAssets struct {
	APIServerCACert    string
	APIServerKey       string
	APIServerCert      string
	CalicoClientCACert string
	CalicoClientKey    string
	CalicoClientCert   string
	EtcdServerCACert   string
	EtcdServerKey      string
	EtcdServerCert     string
}

func readRawTLSAssets(tlsAssetsDir string) (*rawTLSAssets, error) {
	r := new(rawTLSAssets)
	files := []struct {
		name          string
		ca, key, cert *[]byte
	}{
		{"apiserver", &r.APIServerCACert, &r.APIServerKey, &r.APIServerCert},
		{"calico/client", &r.CalicoClientCACert, &r.CalicoClientKey, &r.CalicoClientCert},
		{"etcd/server", &r.EtcdServerCACert, &r.EtcdServerKey, &r.EtcdServerCert},
	}
	for _, file := range files {
		caPath := filepath.Join(tlsAssetsDir, file.name+"-ca.pem")
		keyPath := filepath.Join(tlsAssetsDir, file.name+"-key.pem")
		certPath := filepath.Join(tlsAssetsDir, file.name+".pem")

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

func (r *rawTLSAssets) encrypt(awsSession *awssession.Session, kmsKeyARN string) (*encryptedTLSAssets, error) {
	kmsSvc := kms.New(awsSession)
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
		if encryptOutput, err = kmsSvc.Encrypt(&encryptInput); err != nil {
			return []byte{}
		}
		return encryptOutput.CiphertextBlob
	}
	encryptedAssets := encryptedTLSAssets{
		APIServerCACert:    encrypt(r.APIServerCACert),
		APIServerKey:       encrypt(r.APIServerKey),
		APIServerCert:      encrypt(r.APIServerCert),
		CalicoClientCACert: encrypt(r.CalicoClientCACert),
		CalicoClientKey:    encrypt(r.CalicoClientKey),
		CalicoClientCert:   encrypt(r.CalicoClientCert),
		EtcdServerCACert:   encrypt(r.EtcdServerCACert),
		EtcdServerKey:      encrypt(r.EtcdServerKey),
		EtcdServerCert:     encrypt(r.EtcdServerCert),
	}
	if err != nil {
		return nil, microerror.MaskAny(err)
	}
	return &encryptedAssets, nil
}

func (r *encryptedTLSAssets) compact() (*CompactTLSAssets, error) {
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

	compactAssets := CompactTLSAssets{
		APIServerCACert:    compact(r.APIServerCACert),
		APIServerKey:       compact(r.APIServerKey),
		APIServerCert:      compact(r.APIServerCert),
		CalicoClientCACert: compact(r.CalicoClientCACert),
		CalicoClientKey:    compact(r.CalicoClientKey),
		CalicoClientCert:   compact(r.CalicoClientCert),
		EtcdServerCACert:   compact(r.EtcdServerCACert),
		EtcdServerKey:      compact(r.EtcdServerKey),
		EtcdServerCert:     compact(r.EtcdServerCert),
	}

	if err != nil {
		return nil, microerror.MaskAny(err)
	}
	return &compactAssets, nil
}

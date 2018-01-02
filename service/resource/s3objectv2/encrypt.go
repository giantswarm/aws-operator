package s3objectv2

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/giantswarm/randomkeytpr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/certificatetpr"
	"github.com/giantswarm/microerror"
)

type TLSassets struct {
	APIServerCA       []byte
	APIServerKey      []byte
	APIServerCrt      []byte
	WorkerCA          []byte
	WorkerKey         []byte
	WorkerCrt         []byte
	ServiceAccountCA  []byte
	ServiceAccountKey []byte
	ServiceAccountCrt []byte
	CalicoClientCA    []byte
	CalicoClientKey   []byte
	CalicoClientCrt   []byte
	EtcdServerCA      []byte
	EtcdServerKey     []byte
	EtcdServerCrt     []byte
}

// PEM encoded TLS assets.
type rawTLSAssets TLSassets

// Encrypted PEM encoded TLS assets
type encryptedTLSAssets TLSassets

func createRawTLSAssets(assets certificatetpr.AssetsBundle) *rawTLSAssets {
	// TODO refactor this with a for loop iterating over components and asset types
	return &rawTLSAssets{
		APIServerCA:       assets[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.CA}],
		APIServerCrt:      assets[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.Crt}],
		APIServerKey:      assets[certificatetpr.AssetsBundleKey{certificatetpr.APIComponent, certificatetpr.Key}],
		WorkerCA:          assets[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.CA}],
		WorkerCrt:         assets[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.Crt}],
		WorkerKey:         assets[certificatetpr.AssetsBundleKey{certificatetpr.WorkerComponent, certificatetpr.Key}],
		ServiceAccountCA:  assets[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.CA}],
		ServiceAccountCrt: assets[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.Crt}],
		ServiceAccountKey: assets[certificatetpr.AssetsBundleKey{certificatetpr.ServiceAccountComponent, certificatetpr.Key}],
		EtcdServerCA:      assets[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.CA}],
		EtcdServerCrt:     assets[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Crt}],
		EtcdServerKey:     assets[certificatetpr.AssetsBundleKey{certificatetpr.EtcdComponent, certificatetpr.Key}],
		CalicoClientCA:    assets[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.CA}],
		CalicoClientCrt:   assets[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.Crt}],
		CalicoClientKey:   assets[certificatetpr.AssetsBundleKey{certificatetpr.CalicoComponent, certificatetpr.Key}],
	}
}

type rawKeyAssets map[randomkeytpr.Key][]byte
type encryptedKeyAssets map[randomkeytpr.Key][]byte

func (r rawKeyAssets) encrypt(svc KMSClient, kmsKeyARN string) (*encryptedKeyAssets, error) {

	encryptedAssets := make(encryptedKeyAssets)
	for k, v := range r {
		b, err := encryptor(svc, kmsKeyARN, v)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		encryptedAssets[k] = b
	}

	return &encryptedAssets, nil
}

func (r encryptedKeyAssets) compact() (*randomkeytpr.CompactRandomKeyAssets, error) {
	var err error
	compact := func(data []byte) (r string) {
		if err != nil {
			return ""
		}

		r, err = compactor(data)
		if err != nil {
			return ""
		}

		return r
	}

	compactAssets := randomkeytpr.CompactRandomKeyAssets{
		APIServerEncryptionKey: compact(r[randomkeytpr.EncryptionKey]),
	}

	if err != nil {
		return nil, microerror.Mask(err)
	}
	return &compactAssets, nil
}

func (r *rawTLSAssets) encrypt(svc KMSClient, kmsKeyARN string) (*encryptedTLSAssets, error) {
	var err error
	encrypt := func(data []byte) (b []byte) {
		if err != nil {
			return []byte{}
		}

		b, err = encryptor(svc, kmsKeyARN, data)

		if err != nil {
			return []byte{}
		}
		return b
	}
	encryptedAssets := encryptedTLSAssets{
		APIServerCA:       encrypt(r.APIServerCA),
		APIServerKey:      encrypt(r.APIServerKey),
		APIServerCrt:      encrypt(r.APIServerCrt),
		WorkerCA:          encrypt(r.WorkerCA),
		WorkerCrt:         encrypt(r.WorkerCrt),
		WorkerKey:         encrypt(r.WorkerKey),
		ServiceAccountCA:  encrypt(r.ServiceAccountCA),
		ServiceAccountCrt: encrypt(r.ServiceAccountCrt),
		ServiceAccountKey: encrypt(r.ServiceAccountKey),
		CalicoClientCA:    encrypt(r.CalicoClientCA),
		CalicoClientKey:   encrypt(r.CalicoClientKey),
		CalicoClientCrt:   encrypt(r.CalicoClientCrt),
		EtcdServerCA:      encrypt(r.EtcdServerCA),
		EtcdServerKey:     encrypt(r.EtcdServerKey),
		EtcdServerCrt:     encrypt(r.EtcdServerCrt),
	}
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return &encryptedAssets, nil
}

func (r *encryptedTLSAssets) compact() (*certificatetpr.CompactTLSAssets, error) {
	var err error
	compact := func(data []byte) (r string) {
		if err != nil {
			return ""
		}

		r, err = compactor(data)
		if err != nil {
			return ""
		}

		return r
	}

	compactAssets := certificatetpr.CompactTLSAssets{
		APIServerCA:       compact(r.APIServerCA),
		APIServerKey:      compact(r.APIServerKey),
		APIServerCrt:      compact(r.APIServerCrt),
		WorkerCA:          compact(r.WorkerCA),
		WorkerKey:         compact(r.WorkerKey),
		WorkerCrt:         compact(r.WorkerCrt),
		ServiceAccountCA:  compact(r.ServiceAccountCA),
		ServiceAccountKey: compact(r.ServiceAccountKey),
		ServiceAccountCrt: compact(r.ServiceAccountCrt),
		CalicoClientCA:    compact(r.CalicoClientCA),
		CalicoClientKey:   compact(r.CalicoClientKey),
		CalicoClientCrt:   compact(r.CalicoClientCrt),
		EtcdServerCA:      compact(r.EtcdServerCA),
		EtcdServerKey:     compact(r.EtcdServerKey),
		EtcdServerCrt:     compact(r.EtcdServerCrt),
	}

	if err != nil {
		return nil, microerror.Mask(err)
	}
	return &compactAssets, nil
}

func compactor(data []byte) (string, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(data); err != nil {
		return "", err
	}
	if err := gzw.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func encryptor(svc KMSClient, kmsKeyARN string, data []byte) ([]byte, error) {
	encryptInput := kms.EncryptInput{
		KeyId:     aws.String(kmsKeyARN),
		Plaintext: data,
	}

	var encryptOutput *kms.EncryptOutput
	var err error
	if encryptOutput, err = svc.Encrypt(&encryptInput); err != nil {
		return []byte{}, err
	}
	return encryptOutput.CiphertextBlob, nil
}

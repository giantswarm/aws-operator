package legacyv2

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/randomkeytpr"
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

func createRawTLSAssets(assets legacy.AssetsBundle) *rawTLSAssets {
	// TODO refactor this with a for loop iterating over components and asset types
	return &rawTLSAssets{
		APIServerCA:       assets[legacy.AssetsBundleKey{legacy.APIComponent, legacy.CA}],
		APIServerCrt:      assets[legacy.AssetsBundleKey{legacy.APIComponent, legacy.Crt}],
		APIServerKey:      assets[legacy.AssetsBundleKey{legacy.APIComponent, legacy.Key}],
		WorkerCA:          assets[legacy.AssetsBundleKey{legacy.WorkerComponent, legacy.CA}],
		WorkerCrt:         assets[legacy.AssetsBundleKey{legacy.WorkerComponent, legacy.Crt}],
		WorkerKey:         assets[legacy.AssetsBundleKey{legacy.WorkerComponent, legacy.Key}],
		ServiceAccountCA:  assets[legacy.AssetsBundleKey{legacy.ServiceAccountComponent, legacy.CA}],
		ServiceAccountCrt: assets[legacy.AssetsBundleKey{legacy.ServiceAccountComponent, legacy.Crt}],
		ServiceAccountKey: assets[legacy.AssetsBundleKey{legacy.ServiceAccountComponent, legacy.Key}],
		EtcdServerCA:      assets[legacy.AssetsBundleKey{legacy.EtcdComponent, legacy.CA}],
		EtcdServerCrt:     assets[legacy.AssetsBundleKey{legacy.EtcdComponent, legacy.Crt}],
		EtcdServerKey:     assets[legacy.AssetsBundleKey{legacy.EtcdComponent, legacy.Key}],
		CalicoClientCA:    assets[legacy.AssetsBundleKey{legacy.CalicoComponent, legacy.CA}],
		CalicoClientCrt:   assets[legacy.AssetsBundleKey{legacy.CalicoComponent, legacy.Crt}],
		CalicoClientKey:   assets[legacy.AssetsBundleKey{legacy.CalicoComponent, legacy.Key}],
	}
}

type rawKeyAssets map[randomkeytpr.Key][]byte
type encryptedKeyAssets map[randomkeytpr.Key][]byte

func (r rawKeyAssets) encrypt(svc *kms.KMS, kmsKeyARN string) (*encryptedKeyAssets, error) {

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

func (r *rawTLSAssets) encrypt(svc *kms.KMS, kmsKeyARN string) (*encryptedTLSAssets, error) {
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

func (r *encryptedTLSAssets) compact() (*legacy.CompactTLSAssets, error) {
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

	compactAssets := legacy.CompactTLSAssets{
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

func encryptor(svc *kms.KMS, kmsKeyARN string, data []byte) ([]byte, error) {
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

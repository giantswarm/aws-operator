package encrypter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"

	"github.com/giantswarm/legacycerts/legacy"
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
type RawTLSAssets TLSassets

// Encrypted PEM encoded TLS assets
type EncryptedTLSAssets TLSassets

func CreateRawTLSAssets(assets legacy.AssetsBundle) *RawTLSAssets {
	// TODO refactor this with a for loop iterating over components and asset types
	return &RawTLSAssets{
		APIServerCA:       assets[assetsBundleKey(legacy.APIComponent, legacy.CA)],
		APIServerCrt:      assets[assetsBundleKey(legacy.APIComponent, legacy.Crt)],
		APIServerKey:      assets[assetsBundleKey(legacy.APIComponent, legacy.Key)],
		WorkerCA:          assets[assetsBundleKey(legacy.WorkerComponent, legacy.CA)],
		WorkerCrt:         assets[assetsBundleKey(legacy.WorkerComponent, legacy.Crt)],
		WorkerKey:         assets[assetsBundleKey(legacy.WorkerComponent, legacy.Key)],
		ServiceAccountCA:  assets[assetsBundleKey(legacy.ServiceAccountComponent, legacy.CA)],
		ServiceAccountCrt: assets[assetsBundleKey(legacy.ServiceAccountComponent, legacy.Crt)],
		ServiceAccountKey: assets[assetsBundleKey(legacy.ServiceAccountComponent, legacy.Key)],
		EtcdServerCA:      assets[assetsBundleKey(legacy.EtcdComponent, legacy.CA)],
		EtcdServerCrt:     assets[assetsBundleKey(legacy.EtcdComponent, legacy.Crt)],
		EtcdServerKey:     assets[assetsBundleKey(legacy.EtcdComponent, legacy.Key)],
		CalicoClientCA:    assets[assetsBundleKey(legacy.CalicoComponent, legacy.CA)],
		CalicoClientCrt:   assets[assetsBundleKey(legacy.CalicoComponent, legacy.Crt)],
		CalicoClientKey:   assets[assetsBundleKey(legacy.CalicoComponent, legacy.Key)],
	}
}

func (r *RawTLSAssets) Encrypt(ctx context.Context, e Encrypter, key string) (*EncryptedTLSAssets, error) {
	var err error
	encrypt := func(data []byte) (b []byte) {
		if err != nil {
			return []byte{}
		}

		bs, err := e.Encrypt(ctx, key, string(data))

		if err != nil {
			return []byte{}
		}
		return []byte(bs)
	}
	encryptedAssets := EncryptedTLSAssets{
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

func (r *EncryptedTLSAssets) Compact() (*legacy.CompactTLSAssets, error) {
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

func assetsBundleKey(c legacy.ClusterComponent, t legacy.TLSAssetType) legacy.AssetsBundleKey {
	return legacy.AssetsBundleKey{
		Component: c,
		Type:      t,
	}
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

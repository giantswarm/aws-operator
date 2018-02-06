package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs/legacy"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeytpr"
)

func Test_Service_CloudConfig_NewMasterTemplate(t *testing.T) {
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		Certs        legacy.CompactTLSAssets
		ClusterKeys  randomkeytpr.CompactRandomKeyAssets
	}{
		{
			CustomObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			Certs: legacy.CompactTLSAssets{
				CalicoClientCA:  "123456789-super-magic-calico-client-ca",
				CalicoClientCrt: "123456789-super-magic-calico-client-crt",
				CalicoClientKey: "123456789-super-magic-calico-client-key",
			},
		},
	}

	for _, tc := range testCases {
		ccService, err := testNewCloudConfigService()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		template, err := ccService.NewMasterTemplate(tc.CustomObject, tc.Certs, tc.ClusterKeys)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		decoded, err := testDecodeTemplate(template)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		t.Run("VerifyAPIServerCA", func(t *testing.T) {
			if !strings.Contains(decoded, tc.Certs.CalicoClientCA) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client CA", "none")
			}
		})

		t.Run("VerifyAPIServerCrt", func(t *testing.T) {
			if !strings.Contains(decoded, tc.Certs.CalicoClientCrt) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Crt", "none")
			}
		})

		t.Run("VerifyAPIServerKey", func(t *testing.T) {
			if !strings.Contains(decoded, tc.Certs.CalicoClientKey) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Key", "none")
			}
		})

		t.Run("VerifyTLSAssetsDecryptionUnit", func(t *testing.T) {
			if !strings.Contains(decoded, "- name: decrypt-tls-assets.service") {
				t.Fatalf("expected %#v got %#v", "cloud config to contain unit decrypt-tls-assets.service", "none")
			}
		})
	}
}

func Test_Service_CloudConfig_NewWorkerTemplate(t *testing.T) {
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		Certs        legacy.CompactTLSAssets
	}{
		{
			CustomObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					AWS: v1alpha1.AWSConfigSpecAWS{
						Region: "123456789-super-magic-aws-region",
					},
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
					},
				},
			},
			Certs: legacy.CompactTLSAssets{
				CalicoClientCA:  "123456789-super-magic-calico-client-ca",
				CalicoClientCrt: "123456789-super-magic-calico-client-crt",
				CalicoClientKey: "123456789-super-magic-calico-client-key",
			},
		},
	}

	for _, tc := range testCases {
		ccService, err := testNewCloudConfigService()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		template, err := ccService.NewWorkerTemplate(tc.CustomObject, tc.Certs)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		decoded, err := testDecodeTemplate(template)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		t.Run("VerifyAPIServerCA", func(t *testing.T) {
			if !strings.Contains(decoded, tc.Certs.CalicoClientCA) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client CA", "none")
			}
		})

		t.Run("VerifyAPIServerCrt", func(t *testing.T) {
			if !strings.Contains(decoded, tc.Certs.CalicoClientCrt) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Crt", "none")
			}
		})

		t.Run("VerifyAPIServerKey", func(t *testing.T) {
			if !strings.Contains(decoded, tc.Certs.CalicoClientKey) {
				t.Fatalf("expected %#v got %#v", "cloud config to contain Calico client Key", "none")
			}
		})

		t.Run("VerifyTLSAssetsDecryptionUnit", func(t *testing.T) {
			if !strings.Contains(decoded, "- name: decrypt-tls-assets.service") {
				t.Fatalf("expected %#v got %#v", "cloud config to contain unit decrypt-tls-assets.service", "none")
			}
		})

		t.Run("VerifyAWSRegion", func(t *testing.T) {
			if !strings.Contains(decoded, "--region 123456789-super-magic-aws-region kms decrypt") {
				t.Fatalf("expected %#v got %#v", "cloud config to contain AWS region", "none")
			}
		})
	}
}

func testDecodeTemplate(template string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(template)
	if err != nil {
		return "", err
	}
	r, err := gzip.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, r)
	if err != nil {
		return "", err
	}
	r.Close()

	return b.String(), nil
}

func testNewCloudConfigService() (*CloudConfig, error) {
	var err error

	var ccService *CloudConfig
	{
		ccConfig := DefaultConfig()

		ccConfig.Logger = microloggertest.New()

		ccService, err = New(ccConfig)
		if err != nil {
			return nil, err
		}
	}

	return ccService, nil
}

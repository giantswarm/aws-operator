package cloudconfig

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"io"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v21/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v21/encrypter"
)

func Test_Service_CloudConfig_NewMasterTemplate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		ClusterKeys  randomkeys.Cluster
	}{
		{
			CustomObject: v1alpha1.AWSConfig{
				Spec: v1alpha1.AWSConfigSpec{
					Cluster: v1alpha1.Cluster{
						ID: "al9qy",
						Etcd: v1alpha1.ClusterEtcd{
							Port: 2379,
						},
					},
				},
			},
			ClusterKeys: randomkeys.Cluster{
				APIServerEncryptionKey: randomkeys.RandomKey("fekhfiwoiqhoifhwqefoiqwefoikqhwef"),
			},
		},
	}

	for _, tc := range testCases {
		ctlCtx := controllercontext.Context{}
		ctx := controllercontext.NewContext(context.Background(), ctlCtx)

		ccService, err := testNewCloudConfigService()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		template, err := ccService.NewMasterTemplate(ctx, tc.CustomObject, certs.Cluster{}, tc.ClusterKeys)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		decoded, err := testDecodeTemplate(template)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		expectedStrings := []string{
			"- path: /etc/kubernetes/ssl/etcd/client-ca.pem",
			"- path: /etc/kubernetes/ssl/etcd/client-crt.pem",
			"- path: /etc/kubernetes/ssl/etcd/client-key.pem",
			"- name: decrypt-tls-assets.service",
			"H4sIAAAAAAAA/1SNMQ7CMAxF957CF+jQNSviCuwldYgVYTd2aBQh7o4CVREeLL33pf8T8eLgzF7bWkj4JBzoNswrXVCNhB1s06Bo8lCP5gaAEf6wC0OvWOxDq8pGC+oRzmj+6r/UL2GzH43A8x1dt9MhYW90EDDFQFUoR6EQa8YglGv/KceKYR+hBblQaQ6er3cAAAD//9QjGEbUAAAA",
		}
		for _, expectedString := range expectedStrings {
			if !strings.Contains(decoded, expectedString) {
				t.Fatalf("want decoded to conain %q", expectedString)
			}
		}
	}
}

func Test_Service_CloudConfig_NewWorkerTemplate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
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
		},
	}

	for _, tc := range testCases {
		ctlCtx := controllercontext.Context{}
		ctx := controllercontext.NewContext(context.Background(), ctlCtx)

		ccService, err := testNewCloudConfigService()
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		template, err := ccService.NewWorkerTemplate(ctx, tc.CustomObject, certs.Cluster{})
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		decoded, err := testDecodeTemplate(template)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		expectedStrings := []string{
			"- path: /etc/kubernetes/ssl/etcd/client-ca.pem",
			"- path: /etc/kubernetes/ssl/etcd/client-crt.pem",
			"- path: /etc/kubernetes/ssl/etcd/client-key.pem",
			"- name: decrypt-tls-assets.service",
			"--region 123456789-super-magic-aws-region kms decrypt",
		}
		for _, expectedString := range expectedStrings {
			if !strings.Contains(decoded, expectedString) {
				t.Fatalf("want decoded to conain %q", expectedString)
			}
		}
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
		c := Config{
			Encrypter:      &encrypter.EncrypterMock{},
			Logger:         microloggertest.New(),
			RegistryDomain: "quay.io",
		}

		ccService, err = New(c)
		if err != nil {
			return nil, err
		}
	}

	return ccService, nil
}

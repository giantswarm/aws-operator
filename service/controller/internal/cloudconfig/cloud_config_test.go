package cloudconfig

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/k8scloudconfig/v6/ignition"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v6/pkg/template"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/internal/encrypter"
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

		data := IgnitionTemplateData{
			CustomObject: tc.CustomObject,
			ClusterKeys:  tc.ClusterKeys,
		}
		template, _, err := ccService.NewMasterTemplate(ctx, data)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}
		fmt.Printf("%s", template)
		templateBytes := []byte(template)
		_, err = ignition.ConvertTemplatetoJSON(templateBytes)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		expectedStrings := []string{
			"/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			"/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			"/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			"decrypt-tls-assets.service",
			"PGVuY3J5cHRlZD4tLWtpbmQ6IEVuY3J5cHRpb25Db25maWcKYXBpVmVyc2lvbjogdjEKcmVzb3VyY2VzOgogIC0gcmVzb3VyY2VzOgogICAgLSBzZWNyZXRzCiAgICBwcm92aWRlcnM6CiAgICAtIGFlc2NiYzoKICAgICAgICBrZXlzOgogICAgICAgIC0gbmFtZToga2V5MQogICAgICAgICAgc2VjcmV0OiBmZWtoZml3b2lxaG9pZmh3cWVmb2lxd2Vmb2lrcWh3ZWYKICAgIC0gaWRlbnRpdHk6IHt9",
		}
		for _, expectedString := range expectedStrings {
			if !strings.Contains(template, expectedString) {
				t.Fatalf("want ignition to contain %q", expectedString)
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

		data := IgnitionTemplateData{
			CustomObject: tc.CustomObject,
		}
		template, _, err := ccService.NewWorkerTemplate(ctx, data)
		if err != nil {
			t.Fatalf("expected %#v got %#v", nil, err)
		}

		expectedStrings := []string{
			"/etc/kubernetes/ssl/etcd/client-ca.pem.enc",
			"/etc/kubernetes/ssl/etcd/client-crt.pem.enc",
			"/etc/kubernetes/ssl/etcd/client-key.pem.enc",
			"decrypt-tls-assets.service",
		}
		for _, expectedString := range expectedStrings {
			if !strings.Contains(template, expectedString) {
				t.Fatalf("want ignition to contain %q", expectedString)
			}
		}
	}
}

func testNewCloudConfigService() (*CloudConfig, error) {
	var ccService *CloudConfig
	{
		packagePath, err := k8scloudconfig.GetPackagePath()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		c := Config{
			Encrypter:                 &encrypter.EncrypterMock{},
			Logger:                    microloggertest.New(),
			IgnitionPath:              packagePath,
			ImagePullProgressDeadline: "1m",
			RegistryDomain:            "quay.io",
		}

		ccService, err = New(c)
		if err != nil {
			return nil, err
		}
	}

	return ccService, nil
}

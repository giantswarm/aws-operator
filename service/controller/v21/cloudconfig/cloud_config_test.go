package cloudconfig

import (
	"context"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_0_0"
	"github.com/giantswarm/microerror"
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
			"SDRzSUFBQUFBQUFBLzFTTk1RN0NNQXhGOTU3Q0YralFOU3ZpQ3V3bGRZZ1ZZVGQyYUJRaDdvNENWUkVlTEwzM3BmOFQ4ZUxnekY3YldrajRKQnpvTnN3clhWQ05oQjFzMDZCbzhsQ1A1Z2FBRWY2d0MwT3ZXT3hEcThwR0Mrb1J6bWorNnIvVUwyR3pINDNBOHgxZHQ5TWhZVzkwRURERlFGVW9SNkVRYThZZ2xHdi9LY2VLWVIraEJibFFhUTZlcjNjQUFBRC8vOVFqR0ViVUFBQUE=",
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

		template, err := ccService.NewWorkerTemplate(ctx, tc.CustomObject, certs.Cluster{})
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
			Encrypter:      &encrypter.EncrypterMock{},
			Logger:         microloggertest.New(),
			IgnitionPath:   packagePath,
			RegistryDomain: "quay.io",
		}

		ccService, err = New(c)
		if err != nil {
			return nil, err
		}
	}

	return ccService, nil
}

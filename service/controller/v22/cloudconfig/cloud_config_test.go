package cloudconfig

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
	"github.com/giantswarm/certs"
	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_0_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"

	"github.com/giantswarm/aws-operator/service/controller/v22/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/v22/encrypter"
)

func Test_Service_CloudConfig_NewMasterTemplate(t *testing.T) {
	fixturesCerts, err := testLoadFixturesClusterCerts()
	if err != nil {
		t.Fatalf("error reading fitures %#v", err)
	}

	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		ClusterKeys  randomkeys.Cluster
		ClusterCerts certs.Cluster
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
			ClusterCerts: certs.Cluster{},
		},
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
			ClusterCerts: certs.Cluster{
				APIServer:        fixturesCerts,
				CalicoEtcdClient: fixturesCerts,
				EtcdServer:       fixturesCerts,
				ServiceAccount:   fixturesCerts,
				Worker:           fixturesCerts,
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
		template, err := ccService.NewMasterTemplate(ctx, tc.CustomObject, tc.ClusterCerts, tc.ClusterKeys)
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
			"a2luZDogRW5jcnlwdGlvbkNvbmZpZwphcGlWZXJzaW9uOiB2MQpyZXNvdXJjZXM6CiAgLSByZXNvdXJjZXM6CiAgICAtIHNlY3JldHMKICAgIHByb3ZpZGVyczoKICAgIC0gYWVzY2JjOgogICAgICAgIGtleXM6CiAgICAgICAgLSBuYW1lOiBrZXkxCiAgICAgICAgICBzZWNyZXQ6IGZla2hmaXdvaXFob2lmaHdxZWZvaXF3ZWZvaWtxaHdlZgogICAgLSBpZGVudGl0eToge30=",
		}
		for _, expectedString := range expectedStrings {
			if !strings.Contains(template, expectedString) {
				t.Fatalf("want ignition to contain %q", expectedString)
			}
		}
	}
}

func Test_Service_CloudConfig_NewWorkerTemplate(t *testing.T) {
	fixturesCerts, err := testLoadFixturesClusterCerts()
	if err != nil {
		t.Fatalf("error reading fitures %#v", err)
	}

	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.AWSConfig
		ClusterCerts certs.Cluster
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
			ClusterCerts: certs.Cluster{},
		},
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
			ClusterCerts: certs.Cluster{
				APIServer:        fixturesCerts,
				CalicoEtcdClient: fixturesCerts,
				EtcdServer:       fixturesCerts,
				ServiceAccount:   fixturesCerts,
				Worker:           fixturesCerts,
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

		template, err := ccService.NewWorkerTemplate(ctx, tc.CustomObject, tc.ClusterCerts)
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

func testLoadFixturesClusterCerts() (certs.TLS, error) {
	ca, err := ioutil.ReadFile("testdata/ca.crt")
	if err != nil {
		return certs.TLS{}, err
	}

	crt, err := ioutil.ReadFile("testdata/tls.crt")
	if err != nil {
		return certs.TLS{}, err
	}

	key, err := ioutil.ReadFile("testdata/tls.key")
	if err != nil {
		return certs.TLS{}, err
	}

	return certs.TLS{CA: ca, Crt: crt, Key: key}, nil
}

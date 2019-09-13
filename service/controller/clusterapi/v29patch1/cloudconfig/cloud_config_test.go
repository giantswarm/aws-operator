package cloudconfig

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/giantswarm/certs"
	ignition "github.com/giantswarm/k8scloudconfig/ignition/v_2_2_0"
	k8scloudconfig "github.com/giantswarm/k8scloudconfig/v_4_7_0"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/randomkeys"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"

	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29patch1/controllercontext"
	"github.com/giantswarm/aws-operator/service/controller/clusterapi/v29patch1/encrypter"
)

func Test_Service_CloudConfig_NewMasterTemplate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.Cluster
		ClusterKeys  randomkeys.Cluster
	}{
		{
			CustomObject: v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProviderSpec: v1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: []byte(`
								{
									"cluster": {
										"dns": {
											"domain": "giantswarm.io"
										}
									}
								}
							`),
						},
					},
				},
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "al9qy"
								}
							}
						`),
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
	t.Parallel()
	testCases := []struct {
		CustomObject v1alpha1.Cluster
	}{
		{
			CustomObject: v1alpha1.Cluster{
				Spec: v1alpha1.ClusterSpec{
					ProviderSpec: v1alpha1.ProviderSpec{
						Value: &runtime.RawExtension{
							Raw: []byte(`
								{
									"cluster": {
										"dns": {
											"domain": "giantswarm.io"
										}
									}
								}
							`),
						},
					},
				},
				Status: v1alpha1.ClusterStatus{
					ProviderStatus: &runtime.RawExtension{
						Raw: []byte(`
							{
								"cluster": {
									"id": "al9qy"
								},
								"provider": {
									"region": "123456789-super-magic-aws-region"
								}
							}
						`),
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
			Encrypter: &encrypter.EncrypterMock{},
			Logger:    microloggertest.New(),

			CalicoCIDR:                18,
			CalicoMTU:                 1430,
			CalicoSubnet:              "172.18.128.0",
			ClusterIPRange:            "172.18.192.0/22",
			DockerDaemonCIDR:          "172.18.224.1/19",
			IgnitionPath:              packagePath,
			ImagePullProgressDeadline: "1m",
			NetworkSetupDockerImage:   "quay.io/giantswarm/k8s-setup-network-environment",
			RegistryDomain:            "quay.io",
			SSHUserList:               "user:ssh-rsa base64==",
			SSOPublicKey:              "user:ssh-rsa base64==",
		}

		ccService, err = New(c)
		if err != nil {
			return nil, err
		}
	}

	return ccService, nil
}

package v_3_0_0

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/giantswarm/apiextensions/pkg/apis/provider/v1alpha1"
)

func TestCloudConfig(t *testing.T) {
	tests := []struct {
		template         string
		params           Params
		expectedEtcdPort int
	}{
		{
			template: MasterTemplate,
			params: Params{
				Extension: nopExtension{},
			},
			expectedEtcdPort: 443,
		},
		{
			template: WorkerTemplate,
			params: Params{
				Extension: nopExtension{},
			},
			expectedEtcdPort: 443,
		},
		{
			template: WorkerTemplate,
			params: Params{
				EtcdPort:  2379,
				Extension: nopExtension{},
			},
			expectedEtcdPort: 2379,
		},
	}

	for _, tc := range tests {
		c := DefaultCloudConfigConfig()

		c.Params = tc.params
		c.Template = tc.template

		cloudConfig, err := NewCloudConfig(c)
		if err != nil {
			t.Fatal(err)
		}

		if cloudConfig.params.EtcdPort != tc.expectedEtcdPort {
			t.Errorf("expected etcd port %q, got %q", tc.expectedEtcdPort, cloudConfig.params.EtcdPort)
		}

		if err := cloudConfig.ExecuteTemplate(); err != nil {
			t.Fatal(err)
		}

		cloudConfigBase64 := cloudConfig.Base64()
		if _, err := base64.StdEncoding.DecodeString(cloudConfigBase64); err != nil {
			t.Errorf("The string isn't Base64 encoded: %v", err)
		}
	}
}

func TestCloudConfigTemplating(t *testing.T) {
	tests := []struct {
		template string
		params   Params

		expectedString string
	}{
		// Test that calico range is properly templated.
		{
			template: MasterTemplate,
			params: Params{
				Cluster: v1alpha1.Cluster{
					Calico: v1alpha1.ClusterCalico{
						Subnet: "127.0.0.1",
						CIDR:   16,
					},
				},
				Extension: nopExtension{},
			},

			expectedString: `
                - name: CALICO_IPV4POOL_CIDR
                  value: "127.0.0.1/16"
`,
		},

		// A different test for calico range templating.
		{
			template: MasterTemplate,
			params: Params{
				Cluster: v1alpha1.Cluster{
					Calico: v1alpha1.ClusterCalico{
						Subnet: "192.168.0.0",
						CIDR:   24,
					},
				},
				Extension: nopExtension{},
			},

			expectedString: `
                - name: CALICO_IPV4POOL_CIDR
                  value: "192.168.0.0/24"
`,
		},
	}

	for index, test := range tests {
		c := DefaultCloudConfigConfig()

		c.Params = test.params
		c.Template = test.template

		cloudConfig, err := NewCloudConfig(c)
		if err != nil {
			t.Fatalf("%v: unexpected error creating cloud config: %v", index, err)
		}

		if err := cloudConfig.ExecuteTemplate(); err != nil {
			t.Fatalf("%v: unexpected error templating cloud config: %v", index, err)
		}

		if !strings.Contains(cloudConfig.String(), test.expectedString) {
			t.Fatalf("%v: expected string not found in cloud config: %v", index, test.expectedString)
		}
	}
}

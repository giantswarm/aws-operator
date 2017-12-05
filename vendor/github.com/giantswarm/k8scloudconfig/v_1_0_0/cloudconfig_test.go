package v_1_0_0

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
)

func TestCloudConfig(t *testing.T) {
	tests := []struct {
		template string
		params   Params
	}{
		{
			template: MasterTemplate,
			params: Params{
				Extension: &FakeExtension{},
			},
		},
		{
			template: WorkerTemplate,
			params: Params{
				Extension: &FakeExtension{},
			},
		},
	}

	for _, tc := range tests {
		cloudConfig, err := NewCloudConfig(tc.template, tc.params)
		if err != nil {
			t.Fatal(err)
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
				Cluster: clustertpr.Spec{
					Calico: spec.Calico{
						Subnet: "127.0.0.1",
						CIDR:   16,
					},
				},
				Extension: &FakeExtension{},
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
				Cluster: clustertpr.Spec{
					Calico: spec.Calico{
						Subnet: "192.168.0.0",
						CIDR:   24,
					},
				},
				Extension: &FakeExtension{},
			},

			expectedString: `
                - name: CALICO_IPV4POOL_CIDR
                  value: "192.168.0.0/24"
`,
		},
	}

	for index, test := range tests {
		cloudConfig, err := NewCloudConfig(test.template, test.params)
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

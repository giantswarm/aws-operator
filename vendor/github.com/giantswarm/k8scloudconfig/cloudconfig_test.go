package cloudconfig

import (
	"encoding/base64"
	"testing"
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

package cloudconfig

import (
	"encoding/base64"
	"testing"

	"github.com/giantswarm/clustertpr"
	"github.com/giantswarm/clustertpr/spec"
)

func TestCloudConfig(t *testing.T) {
	params := Params{
		Cluster: clustertpr.Spec{
			Cluster: spec.Cluster{
				ID: "example-cluster",
			},
		},
		Extension: &FakeExtension{},
	}

	tests := []struct {
		template string
		params   Params
	}{
		{
			template: MasterTemplate,
			params:   params,
		},
		{
			template: WorkerTemplate,
			params:   params,
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

package cloudconfig

import (
	"encoding/base64"
	"testing"
)

func TestCloudConfig(t *testing.T) {
	masterTemplate, err := Asset("templates/master.yaml")
	if err != nil {
		t.Fatal(err)
	}
	workerTemplate, err := Asset("templates/worker.yaml")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		template  []byte
		params    CloudConfigTemplateParams
		extension OperatorExtension
	}{
		{
			template:  masterTemplate,
			params:    CloudConfigTemplateParams{},
			extension: &FakeOperatorExtension{},
		},
		{
			template:  workerTemplate,
			params:    CloudConfigTemplateParams{},
			extension: &FakeOperatorExtension{},
		},
	}

	for _, tc := range tests {
		cloudConfig, err := NewCloudConfig(tc.template, tc.params, tc.extension)
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

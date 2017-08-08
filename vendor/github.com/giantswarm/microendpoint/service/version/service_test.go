package version

import (
	"context"
	"reflect"
	"runtime"
	"testing"
)

func Test_Get(t *testing.T) {
	tests := []struct {
		description   string
		gitCommit     string
		name          string
		source        string
		errorExpected bool
		result        Response
	}{
		// Case 1. A valid configuration.
		{
			description:   "test desc",
			gitCommit:     "b6bf741b5c34be4fff51d944f973318d8b078284",
			name:          "api",
			source:        "microkit",
			errorExpected: false,
			result: Response{
				Description: "test desc",
				GitCommit:   "b6bf741b5c34be4fff51d944f973318d8b078284",
				GoVersion:   runtime.Version(),
				Name:        "api",
				OSArch:      runtime.GOOS + "/" + runtime.GOARCH,
				Source:      "microkit",
			},
		},
		// Case 2. Missing git commit.
		{
			description:   "test desc",
			gitCommit:     "",
			name:          "microendpoint",
			source:        "microkit",
			errorExpected: true,
			result:        Response{},
		},
	}

	for i, tc := range tests {
		config := DefaultConfig()
		config.Description = tc.description
		config.GitCommit = tc.gitCommit
		config.Name = tc.name
		config.Source = tc.source

		service, err := New(config)
		if !tc.errorExpected && err != nil {
			t.Fatal("case", i+1, "expected", nil, "got", err)
		}

		if !tc.errorExpected {
			response, err := service.Get(context.TODO(), DefaultRequest())
			if !tc.errorExpected && err != nil {
				t.Fatal("case", i+1, "expected", nil, "got", err)
			}

			if !reflect.DeepEqual(*response, tc.result) {
				t.Fatal("case", i+1, "expected", tc.result, "got", response)
			}
		}
	}
}

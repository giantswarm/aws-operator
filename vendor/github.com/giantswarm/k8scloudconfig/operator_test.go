package cloudconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeParams struct {
	Foo string
}

func TestRenderAssetContent(t *testing.T) {
	tests := []struct {
		assetPath       string
		params          FakeParams
		expectedContent []string
	}{
		{
			assetPath:       "templates/test/test_template.yml",
			params:          FakeParams{Foo: "bar"},
			expectedContent: []string{"foo: bar"},
		},
	}

	for _, tc := range tests {
		rawAssetContent, err := Asset(tc.assetPath)
		if err != nil {
			t.Fatal(err)
		}

		content, err := RenderAssetContent(rawAssetContent, tc.params)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, tc.expectedContent, content, "content should be equal")
	}
}

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
		assetContent    string
		params          FakeParams
		expectedContent []string
	}{
		{
			assetContent:    testTemplate,
			params:          FakeParams{Foo: "bar"},
			expectedContent: []string{"foo: bar"},
		},
	}

	for _, tc := range tests {
		content, err := RenderAssetContent(tc.assetContent, tc.params)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, tc.expectedContent, content, "content should be equal")
	}
}

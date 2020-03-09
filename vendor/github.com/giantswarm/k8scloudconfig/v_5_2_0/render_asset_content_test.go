package v_5_2_0

import (
	"strings"
	"testing"
)

const (
	testTemplate = `foo: {{.Foo}}`
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
		if strings.Join(content, "\n") != strings.Join(tc.expectedContent, "\n") {
			t.Fatalf("expected %#v, got %#v", tc.expectedContent, content)
		}
	}
}

func TestRenderFileAssetContent(t *testing.T) {
	tests := []struct {
		assetContent    string
		params          FakeParams
		expectedContent string
	}{
		{
			assetContent: testTemplate,
			params:       FakeParams{Foo: "bar"},
			// expected base64 encoding of `foo: bar`
			expectedContent: "Zm9vOiBiYXI=",
		},
	}

	for _, tc := range tests {
		content, err := RenderFileAssetContent(tc.assetContent, tc.params)
		if err != nil {
			t.Fatal(err)
		}
		if content != tc.expectedContent {
			t.Fatalf("expected %#v, got %#v", tc.expectedContent, content)
		}
	}
}

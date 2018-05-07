package v_2_0_2

import (
	"bytes"
	"strings"
	"text/template"
)

const (
	GzipBase64 string = "gzip+base64"
)

type FileMetadata struct {
	AssetContent string
	Path         string
	Owner        string
	Encoding     string
	Permissions  int
}

type FileAsset struct {
	Metadata FileMetadata
	Content  []string
}

type UnitMetadata struct {
	AssetContent string
	Name         string
	Enable       bool
	Command      string
}

type UnitAsset struct {
	Metadata UnitMetadata
	Content  []string
}

// VerbatimSection is a blob of YAML we want to add to the
// CloudConfig, with no variable interpolation.
type VerbatimSection struct {
	Name    string
	Content string
}

type Extension interface {
	Files() ([]FileAsset, error)
	Units() ([]UnitAsset, error)
	VerbatimSections() []VerbatimSection
}

func RenderAssetContent(assetContent string, params interface{}) ([]string, error) {
	tmpl, err := template.New("").Parse(assetContent)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, params); err != nil {
		return nil, err
	}

	content := strings.Split(buf.String(), "\n")
	return content, nil
}

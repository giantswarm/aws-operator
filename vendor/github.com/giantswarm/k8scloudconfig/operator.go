package cloudconfig

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

type Extension interface {
	Files() ([]FileAsset, error)
	Units() ([]UnitAsset, error)
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

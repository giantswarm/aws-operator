package cloudconfig

import (
	"bytes"
	"strings"
	"text/template"
)

type encoding string

const (
	UTF8       encoding = "utf-8"
	GzipBase64 encoding = "gzip+base64"
)

func (e encoding) String() string {
	if e == "" {
		return string(UTF8)
	}

	return string(e)
}

type FileMetadata struct {
	AssetContent string
	Path         string
	Owner        string
	Encoding     encoding
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

type OperatorExtension interface {
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

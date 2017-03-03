package cloudconfig

import (
	"bytes"
	"html/template"
	"strings"
)

type FileMetadata struct {
	AssetPath   string
	Path        string
	Owner       string
	Permissions int
}

type FileAsset struct {
	Metadata FileMetadata
	Content  []string
}

type UnitMetadata struct {
	AssetPath string
	Name      string
	Enable    bool
	Command   string
}

type UnitAsset struct {
	Metadata UnitMetadata
	Content  []string
}

type OperatorExtension interface {
	Files() ([]FileAsset, error)
	Units() ([]UnitAsset, error)
}

func RenderAssetContent(rawAssetContent []byte, params interface{}) ([]string, error) {
	tmpl, err := template.New("").Parse(string(rawAssetContent[:]))
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

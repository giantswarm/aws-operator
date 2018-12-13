package v_3_7_4

import (
	"bytes"
	"strings"
	"text/template"
)

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

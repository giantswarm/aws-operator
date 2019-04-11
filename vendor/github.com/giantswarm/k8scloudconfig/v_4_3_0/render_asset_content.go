package v_4_3_0

import (
	"bytes"
	"encoding/base64"
	"strings"
	"text/template"

	"github.com/giantswarm/microerror"
)

func RenderAssetContent(assetContent string, params interface{}) ([]string, error) {
	tmpl, err := template.New("").Parse(assetContent)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, params); err != nil {
		return nil, microerror.Mask(err)
	}

	content := strings.Split(buf.String(), "\n")
	return content, nil
}

// RenderFileAssetContent returns base64 representation of the rendered assetContent.
func RenderFileAssetContent(assetContent string, params interface{}) (string, error) {
	tmpl, err := template.New("").Parse(assetContent)
	if err != nil {
		return "", microerror.Mask(err)
	}

	buf := new(bytes.Buffer)

	if err := tmpl.Execute(buf, params); err != nil {
		return "", microerror.Mask(err)
	}

	content := base64.StdEncoding.EncodeToString(buf.Bytes())
	return content, nil
}

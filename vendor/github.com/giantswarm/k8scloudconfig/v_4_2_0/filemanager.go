package v_4_2_0

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"text/template"

	"github.com/giantswarm/microerror"
)

const (
	version  = "v_4_2_0"
	filesDir = "files"
)

// Files is map[string]string (k: filename, v: contents) for files that are fetched from disk
// and then filled with data.
type Files map[string]string

// RenderFiles walks over filesdir and parses all regular files with
// text/template. Parsed templates are then rendered with ctx, base64 encoded
// and added to returned Files.
//
// filesdir must not contain any other files than templates that can be parsed
// with text/template.
func RenderFiles(filesdir string, ctx interface{}) (Files, error) {
	files := Files{}

	err := filepath.Walk(filesdir, func(path string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				return microerror.Maskf(err, "failed to parse file %#q", path)
			}
			var data bytes.Buffer
			tmpl.Execute(&data, ctx)

			relativePath, err := filepath.Rel(filesdir, path)
			if err != nil {
				return microerror.Mask(err)
			}
			files[relativePath] = base64.StdEncoding.EncodeToString(data.Bytes())
		}
		return nil
	})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return files, nil
}

// GetIgnitionPath returns path for the ignition assets based on
// base ignition directory and package subdirectory with assets.
func GetIgnitionPath(ignitionDir string) string {
	return filepath.Join(ignitionDir, version, filesDir)
}

// GetPackagePath returns top package path for the current runtime file.
// For example, for /go/src/k8scloudconfig/v_4_1_0/file.go function
// returns /go/src/k8scloudconfig.
// This function used only in tests for retrieving ignition assets in runtime.
func GetPackagePath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", microerror.New("failed to retrieve runtime information")
	}

	return filepath.Dir(filepath.Dir(filename)), nil
}

package v_5_2_0

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

const (
	testFilesDir           = "testfiles"
	testConfigDir          = "config"
	testConfigFile         = "config.yaml"
	testConfigFileTemplate = `foo: {{.Foo}}`
)

type RenderParams struct {
	Foo   string
	Files Files
}

func Test_RenderFiles(t *testing.T) {

	// Prepare temporary files structure
	filesDir, err := ioutil.TempDir("", testFilesDir)
	if err != nil {
		t.Fatalf("failed to create temporary directory, %v:", err)
	}
	defer os.RemoveAll(filesDir)

	configPath := path.Join(filesDir, testConfigDir)
	err = os.Mkdir(configPath, 0777)
	if err != nil {
		t.Fatalf("failed to create config directory, %v:", err)
	}

	tmpfile, err := os.Create(path.Join(configPath, testConfigFile))
	if err != nil {
		t.Fatalf("failed to create temporary config file, %v:", err)
	}

	contentBytes := []byte(testConfigFileTemplate)
	if _, err := tmpfile.Write(contentBytes); err != nil {
		t.Fatalf("failed to write template content into temporary file, %v:", err)
	}

	tests := []struct {
		fileContent     string
		filesDir        string
		configDir       string
		configFile      string
		params          RenderParams
		expectedContent string
	}{
		{
			fileContent: testTemplate,
			filesDir:    filesDir,
			configDir:   testConfigDir,
			configFile:  testConfigFile,
			params:      RenderParams{Foo: "bar"},
			// base64 encoded `foo: bar` string
			expectedContent: "Zm9vOiBiYXI=",
		},
	}

	for _, tc := range tests {
		files, err := RenderFiles(tc.filesDir, tc.params)
		if err != nil {
			t.Fatal(err)
		}
		tc.params.Files = files
		contentTemplate := fmt.Sprintf("{{  index .Files \"%s/%s\" }}", tc.configDir, tc.configFile)
		tmpl, err := template.New("").Parse(contentTemplate)
		if err != nil {
			t.Fatal(err)
		}

		buf := new(bytes.Buffer)
		if err := tmpl.Execute(buf, tc.params); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, tc.expectedContent, buf.String(), "content should be equal")
	}
}

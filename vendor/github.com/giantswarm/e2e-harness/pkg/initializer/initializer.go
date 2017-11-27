package initializer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

const (
	ProjectYamlContent = `version: 1

test:
  env:
    - "EXPECTED_KEY=expected_value"
    - "TEST_USER=${USER}"
`
	ClientGoContent = `// +build e2e

package e2e

import (
	"github.com/giantswarm/microerror"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/giantswarm/e2e-harness/pkg/harness"
)

func getK8sClient() (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", harness.DefaultKubeConfig)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return cs, nil
}
`
	ExampleTestGoContent = `// +build e2e

package e2e

import (
	"os"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestZeroInitialPods(t *testing.T) {
	cs, err := getK8sClient()
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	pods, err := cs.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}

	if len(pods.Items) != 0 {
		t.Errorf("Unexpected number of pods, expected 0, got %d", len(pods.Items))
	}
}

func TestEnvVars(t *testing.T) {
	expected := "expected_value"
	actual := os.Getenv("EXPECTED_KEY")

	if expected != actual {
		t.Errorf("unexpected value for EXPECTED_KEY, expected %q, got %q", expected, actual)
	}
}
`
)

type fileDef struct {
	name    string
	content string
}

type Initializer struct {
	logger      micrologger.Logger
	fs          afero.Fs
	projectName string
}

func New(logger micrologger.Logger, fs afero.Fs, projectName string) *Initializer {
	return &Initializer{
		logger:      logger,
		fs:          fs,
		projectName: projectName,
	}
}

func (i *Initializer) CreateLayout() error {
	wd, err := os.Getwd()
	if err != nil {
		return microerror.Mask(err)
	}

	baseDir := filepath.Join(wd, harness.DefaultKubeConfig)

	// return if base dir already exists.
	if _, err := i.fs.Stat(baseDir); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", baseDir)
	}

	if err := i.fs.MkdirAll(baseDir, os.ModePerm); err != nil {
		return microerror.Mask(err)
	}

	afs := &afero.Afero{Fs: i.fs}

	files := []fileDef{
		{
			name:    "project.yaml",
			content: ProjectYamlContent,
		},
		{
			name:    "client.go",
			content: ClientGoContent,
		},
		{
			name:    "example_test.go",
			content: ExampleTestGoContent,
		},
	}

	if err := i.writeFiles(files, baseDir, afs); err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (i *Initializer) writeFiles(files []fileDef, baseDir string, afs *afero.Afero) error {
	for _, f := range files {
		path := filepath.Join(baseDir, f.name)

		if err := afs.WriteFile(path, []byte(f.content), os.ModePerm); err != nil {
			return microerror.Mask(err)
		}
	}
	return nil
}

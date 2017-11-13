package initializer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
)

const (
	DockerFileFmt = `FROM alpine:3.6

ADD ./%s-e2e /e2e

ENTRYPOINT ["/e2e"]
`
	MainGoContentFmt = `package main

import (
	"log"
	"os"

	"github.com/giantswarm/%s/e2e/tests"
)

func main() {
	// architect requires this.
	if len(os.Args) > 1 {
		return
	}

	if err := tests.Run(); err != nil {
		log.Println("error running tests: ", err.Error())
		os.Exit(1)
	}
}
`
	ProjectYamlContent = `version: 1

setup:
  - run: kubectl create namespace giantswarm
    waitFor:
      run: kubectl get namespace
      match: giantswarm\s*Active

outOfClusterTests:
  - run: helm registry install quay.io/giantswarm/cert-operator-lab-chart -- -n cert-operator-lab --set imageTag=latest --set clusterName=my-cluster --wait
    waitFor:
      run: kubectl get certificate
      match: No resources found\.

teardown:
  - run: helm delete cert-operator-lab --purge

inClusterTest:
  enabled: true
`
	RunnerGoContent = `package tests

import (
	"github.com/giantswarm/e2e-harness/pkg/results"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/afero"
	"k8s.io/client-go/kubernetes"
)

// Test is a generic type for all the test functions, it returns a
// description of the tests and the eventually returned error.
type Test func() (string, error)

type TestSet struct {
	clientset kubernetes.Interface
	logger    micrologger.Logger
}

var (
	// tests holds the array of functions to be executed.
	tests = []Test{}
	// ts is the test suite that will keep the results.
	ts = &results.TestSuite{}
)

// Run executes all the tests and saves the results.
func Run() error {
	for _, test := range tests {
		ts.Tests++
		desc, err := test()
		tc := results.TestCase{
			Name: desc,
		}
		if err != nil {
			ts.Failures++
			tc.Failure = &results.TestFailure{
				Value: err.Error(),
			}
		}
		ts.TestCases = append(ts.TestCases, tc)
	}
	fs := afero.NewOsFs()
	return results.Write(fs, ts)
}

// Add appends the given test to the existing bundle.
func Add(t Test) {
	tests = append(tests, t)
}
`
	ExampleGoContent = `package tests

import (
	"fmt"

	"github.com/giantswarm/operatorkit/client/k8sclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	Add(TestExample)
}

func TestExample() (string, error) {
  // the description of the test will be shown in the summary.
	desc := "example in-cluster test"

	cs, err := k8sclient.New(k8sclient.DefaultConfig())
	if err != nil {
		return desc, err
	}

  pods, err := cs.CoreV1().Pods("").List(metav1.ListOptions{})
  if err != nil {
    return desc, err
  }

  if len(pods.Items) != 0 {
    return desc, fmt.Errorf("Unexpected number of pods, expected 0, got %d", len(pods.Items))
  }

	return desc, nil
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
		return err
	}

	baseDir := filepath.Join(wd, "e2e")

	// return if base dir already exists.
	if _, err := i.fs.Stat(baseDir); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", baseDir)
	}

	if err := i.fs.MkdirAll(baseDir, os.ModePerm); err != nil {
		return err
	}

	afs := &afero.Afero{Fs: i.fs}

	files := []fileDef{
		{
			name:    "Dockerfile",
			content: fmt.Sprintf(DockerFileFmt, i.projectName),
		},
		{
			name:    "main.go",
			content: fmt.Sprintf(MainGoContentFmt, i.projectName),
		},
		{
			name:    "project.yaml",
			content: ProjectYamlContent,
		},
	}

	if err := i.writeFiles(files, baseDir, afs); err != nil {
		return err
	}

	testsDir := filepath.Join(baseDir, "tests")

	if err := i.fs.MkdirAll(testsDir, os.ModePerm); err != nil {
		return err
	}

	files = []fileDef{
		{
			name:    "runner.go",
			content: RunnerGoContent,
		},
		{
			name:    "example.go",
			content: ExampleGoContent,
		},
	}

	if err := i.writeFiles(files, testsDir, afs); err != nil {
		return err
	}

	return nil
}

func (i *Initializer) writeFiles(files []fileDef, baseDir string, afs *afero.Afero) error {
	for _, f := range files {
		path := filepath.Join(baseDir, f.name)

		if err := afs.WriteFile(path, []byte(f.content), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

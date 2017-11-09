package initializer_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/giantswarm/e2e-harness/pkg/initializer"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/spf13/afero"
)

const (
	projectName = "myproject"
)

func TestCreateLayout(t *testing.T) {
	fs := afero.NewMemMapFs()
	logger := microloggertest.New()

	subject := initializer.New(logger, fs, projectName)

	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("could not get current directory, %v", err)
	}
	baseDir := filepath.Join(wd, "e2e")

	t.Run("fails if base directory exists", func(t *testing.T) {
		if err := fs.MkdirAll(baseDir, os.ModePerm); err != nil {
			t.Errorf("could not create container dir, %v", err)
		}

		if err := subject.CreateLayout(); err == nil {
			t.Errorf("expected error creating layout did not happen")
		}
	})

	fs.RemoveAll(baseDir)
	if err := subject.CreateLayout(); err != nil {
		t.Errorf("unexpected error creating layout %s", err)
	}

	t.Run("creates base directory", func(t *testing.T) {
		if _, err := fs.Stat(baseDir); os.IsNotExist(err) {
			t.Errorf("directory %s was not created", baseDir)
		}
	})

	afs := &afero.Afero{Fs: fs}

	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "Dockerfile",
			expected: fmt.Sprintf(initializer.DockerFileFmt, projectName),
		},
		{
			name:     "main.go",
			expected: fmt.Sprintf(initializer.MainGoContentFmt, projectName),
		},
		{
			name:     "project.yaml",
			expected: initializer.ProjectYamlContent,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			targetFile := filepath.Join(baseDir, tc.name)
			t.Run("is created", func(t *testing.T) {
				if _, err := fs.Stat(targetFile); os.IsNotExist(err) {
					t.Errorf("%s was not created", targetFile)
				}
			})

			t.Run("has the right content", func(t *testing.T) {
				actual, err := afs.ReadFile(targetFile)
				if err != nil {
					t.Errorf("could not read %s, %v", tc.name, err)
				}

				if string(actual) != tc.expected {
					t.Errorf("Wrong %s contents, expected %s, actual %s ", tc.name, tc.expected, actual)
				}
			})
		})

		testsDir := filepath.Join(baseDir, "tests")
		t.Run("creates tests directory", func(t *testing.T) {
			if _, err := fs.Stat(testsDir); os.IsNotExist(err) {
				t.Errorf("directory %s was not created", testsDir)
			}
		})
		testCases := []struct {
			name     string
			expected string
		}{
			{
				name:     "example.go",
				expected: initializer.ExampleGoContent,
			},
			{
				name:     "runner.go",
				expected: initializer.RunnerGoContent,
			},
		}
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				targetFile := filepath.Join(testsDir, tc.name)
				t.Run("is created", func(t *testing.T) {
					if _, err := fs.Stat(targetFile); os.IsNotExist(err) {
						t.Errorf("%s was not created", targetFile)
					}
				})

				t.Run("has the right content", func(t *testing.T) {
					actual, err := afs.ReadFile(targetFile)
					if err != nil {
						t.Errorf("could not read %s, %v", tc.name, err)
					}

					if string(actual) != tc.expected {
						t.Errorf("Wrong %s contents, expected %s, actual %s ", tc.name, tc.expected, actual)
					}
				})
			})
		}
	}
}

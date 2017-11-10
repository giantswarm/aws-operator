package harness

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/giantswarm/micrologger"
	yaml "gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = "config.yaml"
)

type Harness struct {
	logger micrologger.Logger
	cfg    Config
}

type Config struct {
	RemoteCluster bool `yaml:"remoteCluster"`
}

func New(logger micrologger.Logger, cfg Config) *Harness {
	return &Harness{
		logger: logger,
		cfg:    cfg,
	}
}

// Init initializes the harness.
func (h *Harness) Init() error {
	baseDir, err := BaseDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(filepath.Join(baseDir, "workdir"), os.ModePerm)
}

// WriteConfig is a Task that persists the current config to a file.
func (h *Harness) WriteConfig() error {
	dir, err := BaseDir()
	if err != nil {
		return err
	}

	content, err := yaml.Marshal(&h.cfg)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(dir, defaultConfigFile), []byte(content), 0644)

	return err
}

// ReadConfig is a Task that populates a Config struct data read
// from a default file location.
func (h *Harness) ReadConfig() (Config, error) {
	dir, err := BaseDir()
	if err != nil {
		return Config{}, err
	}

	content, err := ioutil.ReadFile(filepath.Join(dir, defaultConfigFile))
	if err != nil {
		return Config{}, err
	}

	c := &Config{}

	if err := yaml.Unmarshal(content, c); err != nil {
		return Config{}, err
	}

	return *c, nil
}

func BaseDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".e2e-harness"), nil
}

func GetProjectName() string {
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "" {
		return os.Getenv("CIRCLE_PROJECT_REPONAME")
	}
	dir, err := os.Getwd()
	if err != nil {
		return "e2e-harness"
	}
	return filepath.Base(dir)
}

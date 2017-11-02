package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/giantswarm/e2e-harness/pkg/harness"
	"github.com/giantswarm/e2e-harness/pkg/results"
	"github.com/giantswarm/e2e-harness/pkg/runner"
	"github.com/giantswarm/e2e-harness/pkg/tasks"
	"github.com/giantswarm/e2e-harness/pkg/wait"
	"github.com/giantswarm/micrologger"
	yaml "gopkg.in/yaml.v2"
)

const (
	DefaultSonobuoyValuesFile = "/workdir/sonobuoy.yaml"
)

type E2e struct {
	Version          string        `yaml:"version"`
	Setup            []Step        `yaml:"setup"`
	Teardown         []Step        `yaml:"teardown"`
	OutOfClusterTest []Step        `yaml:"outOfClusterTest"`
	InClusterTest    InClusterTest `yaml:"inClusterTest"`
}

type Step struct {
	Run     string   `yaml:"run"`
	WaitFor WaitStep `yaml:"waitFor"`
}

type WaitStep struct {
	Run     string        `yaml:"run"`
	Match   string        `yaml:"match"`
	Timeout time.Duration `yaml:"timeout"`
	Step    time.Duration `yaml:"step"`
}

type InClusterTest struct {
	Enabled bool     `yaml:"enabled"`
	Env     []EnvVar `yaml:"env"`
}

type EnvVar struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Config struct {
	Name      string
	GitCommit string
}

type Dependencies struct {
	Logger  micrologger.Logger
	Runner  runner.Runner
	Wait    *wait.Wait
	Results *results.Results
}

type Project struct {
	logger  micrologger.Logger
	runner  runner.Runner
	wait    *wait.Wait
	results *results.Results
	cfg     *Config
}

type SonoBuoyValues struct {
	ImageName string   `yaml:"imageName"`
	ImageTag  string   `yaml:"imageTag"`
	Env       []EnvVar `yaml:"env"`
}

func New(deps *Dependencies, cfg *Config) *Project {
	return &Project{
		logger:  deps.Logger,
		runner:  deps.Runner,
		wait:    deps.Wait,
		results: deps.Results,
		cfg:     cfg,
	}
}

func (p *Project) CommonSetupSteps() error {
	p.logger.Log("info", "executing common setup steps")
	steps := []Step{
		Step{
			Run: "kubectl create clusterrolebinding permissive-binding --clusterrole cluster-admin --group=system:serviceaccounts",
		},
		Step{
			Run: "kubectl -n kube-system create sa tiller",
		},
		Step{
			Run: "kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller",
		},
		Step{
			Run: "helm init --service-account tiller",
			WaitFor: WaitStep{
				Run:   "kubectl get pod -n kube-system",
				Match: `tiller-deploy.*1/1\s*Running`,
			},
		}}

	for _, s := range steps {
		if err := p.runStep(s); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) SetupSteps() error {
	p.logger.Log("info", "executing setup steps")

	e2e, err := p.readProjectFile()
	if err != nil {
		return err
	}

	for _, step := range e2e.Setup {
		if err := p.runStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) TeardownSteps() error {
	p.logger.Log("info", "executing teardown steps")

	e2e, err := p.readProjectFile()
	if err != nil {
		return err
	}

	for _, step := range e2e.Teardown {
		if err := p.runStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) OutOfClusterTest() error {
	p.logger.Log("info", "executing out of cluster tests")

	e2e, err := p.readProjectFile()
	if err != nil {
		return err
	}

	for _, step := range e2e.OutOfClusterTest {
		if err := p.runStep(step); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) InClusterTest() error {
	p.logger.Log("info", "executing in-cluster tests")

	e2e, err := p.readProjectFile()
	if err != nil {
		return err
	}

	if !e2e.InClusterTest.Enabled {
		p.logger.Log("info", "in-cluster tests disabled")
		return nil
	}

	bundle := []tasks.Task{
		p.createSonobuoyValues,
		p.installSonobuoyChart,
		p.results.Unpack,
		p.results.Interpret,
	}
	return tasks.Run(bundle)
}

func (p *Project) runStep(step Step) error {
	p.logger.Log("info", fmt.Sprintf("executing step with command %q", step.Run))
	// expand env vars
	sEnv := os.ExpandEnv(step.Run)
	if err := p.runner.RunPortForward(os.Stdout, sEnv); err != nil {
		return err
	}

	md := &wait.MatchDef{
		Run:      step.WaitFor.Run,
		Match:    step.WaitFor.Match,
		Deadline: step.WaitFor.Timeout,
	}
	if err := p.wait.For(md); err != nil {
		return err
	}
	return nil
}

func (p *Project) readProjectFile() (*E2e, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	projectFile := filepath.Join(dir, "e2e", "project.yaml")
	if _, err := os.Stat(projectFile); os.IsNotExist(err) {
		return nil, err
	}

	content, err := ioutil.ReadFile(projectFile)
	if err != nil {
		return nil, err
	}

	e2e := &E2e{}

	if err := yaml.Unmarshal(content, e2e); err != nil {
		return nil, err
	}
	return e2e, nil
}

func (p *Project) createSonobuoyValues() error {
	p.logger.Log("info", "creating sonobuoy values")

	e2e, err := p.readProjectFile()
	if err != nil {
		return err
	}

	var name string
	if os.Getenv("CIRCLE_PROJECT_REPONAME") != "" {
		name = os.Getenv("CIRCLE_PROJECT_REPONAME")
	} else {
		name = p.cfg.Name
	}
	var tag string
	if os.Getenv("CIRCLE_SHA1") != "" {
		tag = os.Getenv("CIRCLE_SHA1")
	} else {
		tag = p.cfg.GitCommit
	}

	sonobuoyValues := &SonoBuoyValues{
		ImageName: "quay.io/giantswarm/" + name + "-e2e",
		ImageTag:  tag,
		Env:       e2e.InClusterTest.Env,
	}

	content, err := yaml.Marshal(sonobuoyValues)
	if err != nil {
		return err
	}

	basedir, err := harness.BaseDir()
	if err != nil {
		return err
	}
	path := filepath.Join(basedir, DefaultSonobuoyValuesFile)
	err = ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) installSonobuoyChart() error {
	p.logger.Log("info", "cleaning up previous sonobuoy deployments")

	p.runStep(Step{
		Run: "helm delete --purge sonobuoy-chart",
		WaitFor: WaitStep{
			Run:   "kubectl get ns heptio-sonobuoy",
			Match: `Error from server \(NotFound\): namespaces "heptio-sonobuoy" not found`,
		},
	})

	p.logger.Log("info", "installing sonobuoy chart")
	installSonobuoyChart := Step{
		Run: "helm install /home/e2e-harness/resources/sonobuoy-chart --name sonobuoy-chart --values " + DefaultSonobuoyValuesFile,
		WaitFor: WaitStep{
			Run:   "kubectl logs sonobuoy --namespace=heptio-sonobuoy",
			Match: "no-exit was specified, sonobuoy is now blocking",
		},
	}
	if err := p.runStep(installSonobuoyChart); err != nil {
		return err
	}
	return nil
}

// Package version implements the version command for any microservice.
package version

import (
	"fmt"
	"runtime"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/versionbundle"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Description    string
	GitCommit      string
	Name           string
	Source         string
	VersionBundles []versionbundle.Bundle
}

func New(config Config) (Command, error) {
	if config.Description == "" {
		return nil, microerror.Maskf(invalidConfigError, "description commit must not be empty")
	}
	if config.GitCommit == "" {
		return nil, microerror.Maskf(invalidConfigError, "git commit must not be empty")
	}
	if config.Name == "" {
		return nil, microerror.Maskf(invalidConfigError, "name must not be empty")
	}
	if config.Source == "" {
		return nil, microerror.Maskf(invalidConfigError, "source must not be empty")
	}

	newCommand := &command{
		cobraCommand: nil,

		Description:    config.Description,
		GitCommit:      config.GitCommit,
		Name:           config.Name,
		Source:         config.Source,
		GoVersion:      runtime.Version(),
		OS:             runtime.GOOS,
		Arch:           runtime.GOARCH,
		VersionBundles: config.VersionBundles,
	}

	newCommand.cobraCommand = &cobra.Command{
		Use:   "version",
		Short: "Show version information of the microservice.",
		Long:  "Show version information of the microservice.",
		Run:   newCommand.Execute,
	}

	return newCommand, nil
}

type command struct {
	cobraCommand *cobra.Command

	Description    string                 `json:"description" yaml:"description"`
	GitCommit      string                 `json:"gitCommit" yaml:"gitCommit"`
	Name           string                 `json:"name" yaml:"name"`
	Source         string                 `json:"source" yaml:"source"`
	GoVersion      string                 `json:"goVersion" yaml:"goVersion"`
	OS             string                 `json:"os" yaml:"os"`
	Arch           string                 `json:"arch" yaml:"arch"`
	VersionBundles []versionbundle.Bundle `json:"versionBundles" yaml:"versionBundles"`
}

func (c *command) CobraCommand() *cobra.Command {
	return c.cobraCommand
}

func (c *command) Execute(cmd *cobra.Command, args []string) {
	d, err := yaml.Marshal(c)
	if err != nil {
		fmt.Printf("Could not format version data: #%v", err)
	}

	fmt.Printf("%s", d)
}

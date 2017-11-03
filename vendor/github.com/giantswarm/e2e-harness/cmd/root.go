package cmd

import (
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{
		Use:   "e2e-harness",
		Short: "Harness for custom kubernetes e2e testing",
	}

	gitCommit   string
	projectName string
)

func SetGitCommit(value string) {
	gitCommit = value
}

func GetGitCommit() string {
	return gitCommit
}

func SetProjectName(value string) {
	projectName = value
}

func GetProjectName() string {
	return projectName
}

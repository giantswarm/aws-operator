package daemon

import (
	"github.com/spf13/cobra"
)

// Command represents the daemon command for any microservice.
type Command interface {
	// CobraCommand returns the actual cobra command for the daemon command.
	CobraCommand() *cobra.Command
	// Execute represents the cobra run method.
	Execute(cmd *cobra.Command, args []string)
}

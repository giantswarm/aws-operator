package runner

import "io"

type Runner interface {
	Run(out io.Writer, command string, env ...string) error
	RunPortForward(out io.Writer, command string, env ...string) error
}

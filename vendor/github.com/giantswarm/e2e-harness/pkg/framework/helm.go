package framework

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// HelmCmd executes a helm command.
func HelmCmd(cmd string) error {
	return runCmd("helm " + cmd)
}

func runCmd(cmdStr string) error {
	log.Printf("Running command %q\n", cmdStr)
	cmdEnv := os.ExpandEnv(cmdStr)
	fields := strings.Fields(cmdEnv)
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd.Run()
}

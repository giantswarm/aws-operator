// +build k8srequired

package integration

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
)

const (
	defaultTimeout = 400
)

func runCmd(cmdStr string) error {
	log.Printf("Running command %v\n", cmdStr)
	cmdEnv := os.ExpandEnv(cmdStr)
	fields := strings.Fields(cmdEnv)
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	return cmd.Run()
}

func waitFor(f func() error) error {
	timeout := time.After(defaultTimeout * time.Second)
	ticker := backoff.NewTicker(backoff.NewExponentialBackOff())

	for {
		select {
		case <-timeout:
			ticker.Stop()
			return microerror.Mask(waitTimeoutError)
		case <-ticker.C:
			if err := f(); err == nil {
				return nil
			}
		}
	}
}

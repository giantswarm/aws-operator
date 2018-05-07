package framework

import (
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
)

const (
	defaultTimeout = 1800
)

func waitFor(f func() error) error {
	return baseWait(backoff.NewExponentialBackOff(), f)
}

func waitConstantFor(f func() error) error {
	return baseWait(backoff.NewConstantBackOff(1*time.Second), f)
}

func baseWait(b backoff.BackOff, f func() error) error {
	timeout := time.After(defaultTimeout * time.Second)
	ticker := backoff.NewTicker(b)

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

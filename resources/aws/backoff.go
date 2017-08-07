package aws

import (
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/micrologger"
)

const maxElapsedTime = 2 * time.Minute

// NewCustomExponentialBackoff returns pointer to a backoff.ExponentialBackOff,
// initialized with custom values. At the moment, we only override the
// MaxElapsedTime.
func NewCustomExponentialBackoff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      maxElapsedTime,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	return b
}

func NewNotify(logger micrologger.Logger, operationName string) func(error, time.Duration) {
	return func(err error, delay time.Duration) {
		logger.Log("error", fmt.Sprintf("%s failed, retrying with delay %.0fm%.0fs: '%#v'", operationName, delay.Minutes(), delay.Seconds(), err))
	}
}

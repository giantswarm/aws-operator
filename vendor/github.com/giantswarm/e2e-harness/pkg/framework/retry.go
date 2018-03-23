package framework

import (
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
)

const maxElapsedTime = 2 * time.Minute

// newCustomExponentialBackoff returns pointer to a backoff.ExponentialBackOff,
// initialized with custom values. At the moment, we only override the
// maxElapsedTime.
func newCustomExponentialBackoff() *backoff.ExponentialBackOff {
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

func newNotify(operationName string) func(error, time.Duration) {
	return func(err error, delay time.Duration) {
		log.Printf(fmt.Sprintf("%s failed, retrying with delay %.0fm%.0fs: '%#v'", operationName, delay.Minutes(), delay.Seconds(), err))
	}
}

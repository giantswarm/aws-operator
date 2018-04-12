package helmclient

import (
	"time"

	"github.com/cenkalti/backoff"
)

func newExponentialBackoff(maxWait time.Duration) *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoff.DefaultMaxInterval,
		MaxElapsedTime:      maxWait,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	return b
}

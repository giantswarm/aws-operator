package framework

import (
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
)

const (
	LongMaxWait  = 40 * time.Minute
	ShortMaxWait = 2 * time.Minute
)

const (
	LongMaxInterval  = 60 * time.Second
	ShortMaxInterval = 5 * time.Second
)

func NewExponentialBackoff(maxWait, maxInterval time.Duration) *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     backoff.DefaultInitialInterval,
		RandomizationFactor: backoff.DefaultRandomizationFactor,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         maxInterval,
		MaxElapsedTime:      maxWait,
		Clock:               backoff.SystemClock,
	}

	b.Reset()

	return b
}

func NewConstantBackoff(maxWait, maxInterval time.Duration) backoff.BackOff {
	b := WithMaxElapsedTime(backoff.NewConstantBackOff(maxInterval), maxWait)

	b.Reset()

	return b
}

func WithMaxElapsedTime(b backoff.BackOff, d time.Duration) *BackOffMaxElapsedTime {
	return &BackOffMaxElapsedTime{
		delegate:   b,
		maxElapsed: d,
		start:      time.Time{},
	}
}

type BackOffMaxElapsedTime struct {
	delegate   backoff.BackOff
	maxElapsed time.Duration
	start      time.Time
}

func (b *BackOffMaxElapsedTime) NextBackOff() time.Duration {
	if b.start.IsZero() {
		b.start = time.Now()
	}

	if time.Now().After(b.start.Add(b.maxElapsed)) {
		return backoff.Stop
	}

	return b.delegate.NextBackOff()
}

func (b *BackOffMaxElapsedTime) Reset() {
	b.start = time.Time{}
	b.delegate.Reset()
}

func newNotify(operationName string) func(error, time.Duration) {
	return func(err error, delay time.Duration) {
		log.Printf(fmt.Sprintf("%s failed, retrying with delay %.0fm%.0fs: '%#v'", operationName, delay.Minutes(), delay.Seconds(), err))
	}
}

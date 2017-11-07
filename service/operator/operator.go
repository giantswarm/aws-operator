package operator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
	"github.com/giantswarm/operatorkit/tpr"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	BackOff   backoff.BackOff
	Framework *framework.Framework
	Informer  *informer.Informer
	Logger    micrologger.Logger
	TPR       *tpr.TPR
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		BackOff:   nil,
		Framework: nil,
		Informer:  nil,
		Logger:    nil,
		TPR:       nil,
	}
}

// Operator implements the reconciliation of custom objects.
type Operator struct {
	// Dependencies.
	backOff   backoff.BackOff
	framework *framework.Framework
	informer  *informer.Informer
	logger    micrologger.Logger
	tpr       *tpr.TPR

	// Internals.
	bootOnce sync.Once
	mutex    sync.Mutex
}

// New creates a new configured service.
func New(config Config) (*Operator, error) {
	// Dependencies.
	if config.BackOff == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.BackOff must not be empty")
	}
	if config.Framework == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Framework must not be empty")
	}
	if config.Informer == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Informer must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.TPR == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.TPR must not be empty")
	}

	newOperator := &Operator{
		// Dependencies.
		backOff:   config.BackOff,
		framework: config.Framework,
		informer:  config.Informer,
		logger:    config.Logger,
		tpr:       config.TPR,

		// Internals
		bootOnce: sync.Once{},
		mutex:    sync.Mutex{},
	}

	return newOperator, nil
}

func (o *Operator) Boot() {
	o.bootOnce.Do(func() {
		operation := func() error {
			err := o.bootWithError()
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		notifier := func(err error, d time.Duration) {
			o.logger.Log("warning", fmt.Sprintf("retrying operator boot due to error: %#v", microerror.Mask(err)))
		}

		err := backoff.RetryNotify(operation, o.backOff, notifier)
		if err != nil {
			o.logger.Log("error", fmt.Sprintf("stop operator boot retries due to too many errors: %#v", microerror.Mask(err)))
			os.Exit(1)
		}
	})
}

func (o *Operator) bootWithError() error {
	err := o.tpr.CreateAndWait()
	if tpr.IsAlreadyExists(err) {
		o.logger.Log("debug", "third party resource already exists")
	} else if err != nil {
		return microerror.Mask(err)
	}

	o.logger.Log("debug", "starting list/watch")

	deleteChan, updateChan, errChan := o.informer.Watch(context.TODO())
	o.framework.ProcessEvents(context.TODO(), deleteChan, updateChan, errChan)

	return nil
}

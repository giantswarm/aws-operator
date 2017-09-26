package operator

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/giantswarm/kvmtpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/tpr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	BackOff           backoff.BackOff
	K8sClient         kubernetes.Interface
	Logger            micrologger.Logger
	OperatorFramework *framework.Framework
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		BackOff:           nil,
		K8sClient:         nil,
		Logger:            nil,
		OperatorFramework: nil,
	}
}

// Operator implements the reconciliation of custom objects.
type Operator struct {
	// Dependencies.
	backOff           backoff.BackOff
	logger            micrologger.Logger
	operatorFramework *framework.Framework

	// Internals.
	bootOnce sync.Once
	mutex    sync.Mutex
	tpr      *tpr.TPR
}

// New creates a new configured service.
func New(config Config) (*Operator, error) {
	// Dependencies.
	if config.BackOff == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.BackOff must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.OperatorFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.OperatorFramework must not be empty")
	}

	var err error

	var newTPR *tpr.TPR
	{
		c := tpr.DefaultConfig()

		c.K8sClient = config.K8sClient
		c.Logger = config.Logger

		c.Description = kvmtpr.Description
		c.Name = kvmtpr.Name
		c.Version = kvmtpr.VersionV1

		newTPR, err = tpr.New(c)
		if err != nil {
			return nil, microerror.Maskf(err, "creating TPR util for "+kvmtpr.Name)
		}
	}

	newOperator := &Operator{
		// Dependencies.
		backOff:           config.BackOff,
		logger:            config.Logger,
		operatorFramework: config.OperatorFramework,

		// Internals
		bootOnce: sync.Once{},
		mutex:    sync.Mutex{},
		tpr:      newTPR,
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

	newResourceEventHandler := o.operatorFramework.NewCacheResourceEventHandler()

	newZeroObjectFactory := &tpr.ZeroObjectFactoryFuncs{
		NewObjectFunc:     func() runtime.Object { return &kvmtpr.CustomObject{} },
		NewObjectListFunc: func() runtime.Object { return &kvmtpr.List{} },
	}

	o.tpr.NewInformer(newResourceEventHandler, newZeroObjectFactory).Run(nil)

	return nil
}

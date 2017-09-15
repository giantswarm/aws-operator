// Package transaction provides transactional primitives to ensure certain
// actions happen only ones.
package transaction

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"

	transactionid "github.com/giantswarm/microkit/transaction/context/id"
)

// DefaultReplayDecoder is the default decoder used to convert persisted trial
// outputs so they can be consumed by replay functions. The underlying type of
// the returned interface value is string.
var DefaultReplayDecoder = func(b []byte) (interface{}, error) {
	return string(b), nil
}

// DefaultTrialEncoder is the default encoder used to convert created trial
// outputs so they can be persisted.
var DefaultTrialEncoder = func(v interface{}) ([]byte, error) {
	if v == nil {
		return nil, nil
	}

	b, ok := v.([]byte)
	if ok {
		return b, nil
	}

	s, ok := v.(string)
	if ok {
		return []byte(s), nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return b, nil
}

// ExecuterConfig represents the configuration used to create a executer.
type ExecuterConfig struct {
	// Dependencies.
	Logger  micrologger.Logger
	Storage microstorage.Storage
}

// DefaultExecuterConfig provides a default configuration to create a new
// executer by best effort.
func DefaultExecuterConfig() ExecuterConfig {
	return ExecuterConfig{
		// Dependencies.
		Logger:  nil,
		Storage: nil,
	}
}

// NewExecuter creates a new configured executer.
func NewExecuter(config ExecuterConfig) (Executer, error) {
	newExecuter := &executer{
		// Dependencies.
		logger:  config.Logger,
		storage: config.Storage,
	}

	return newExecuter, nil
}

type executer struct {
	// Dependencies.
	logger  micrologger.Logger
	storage microstorage.Storage
}

func (e *executer) Execute(ctx context.Context, config ExecuteConfig) error {
	// Validate the execute config to make sure we can safely work with it.
	err := validateExecuteConfig(config)
	if err != nil {
		return microerror.Mask(err)
	}

	// At first we check for the transaction ID that might be obtained by the
	// given context. We actually do not care about if there is one or not. It
	// will be either set or empty. In case it is set, we use it for the execution
	// of the transaction below. In case it is not set at all, we simply want to
	// execute the configured trial all the time. Note that we also do not keep
	// track of the trial result. There is no transaction ID so we have no
	// reference we could use to track any information reliably.
	transactionID, ok := transactionid.FromContext(ctx)
	if !ok {
		_, err := config.Trial(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		e.logger.Log("debug", fmt.Sprintf("executed transaction trial without transaction ID for trial ID '%s'", config.TrialID))

		return nil
	}

	// Here we know there is a transaction ID given. Thus we want to check if the
	// trial was already successful. If the trial was already successful we want
	// to execute the configured replay, if any. If there was no trial for the
	// given transaction registered yet, we are executing the trial.
	{
		key, err := microstorage.NewK(transactionKey("transaction", transactionID, "trial", config.TrialID))
		if err != nil {
			return microerror.Mask(err)
		}
		exists, err := e.storage.Exists(ctx, key)
		if err != nil {
			return microerror.Mask(err)
		}

		if exists {
			if config.Replay == nil {
				// In case there is no replay function configured, we cannot execute it.
				// Further, the trial function was already executed at this point, so we
				// stop processing the transaction here.
				return nil
			}

			var notFound bool
			key, err := microstorage.NewK(transactionKey("transaction", transactionID, "trial", config.TrialID, "result"))
			if err != nil {
				return microerror.Mask(err)
			}
			kv, err := e.storage.Search(ctx, key)
			if microstorage.IsNotFound(err) {
				notFound = true
			} else if err != nil {
				return microerror.Mask(err)
			}

			// Here it is important to only provide a none nil value to the replay
			// function, if there is really some trial output persisted. This becomes
			// important in cases in which one explicitly expects e.g. empty strings
			// as trial output.
			var input interface{}
			if !notFound {
				input, err = config.ReplayDecoder([]byte(kv.Val()))
				if err != nil {
					return microerror.Mask(err)
				}
			}

			err = config.Replay(ctx, input)
			if err != nil {
				return microerror.Mask(err)
			}
			e.logger.Log("debug", fmt.Sprintf("executed transaction replay for transaction ID '%s' and trial ID '%s'", transactionID, config.TrialID))

			return nil
		}
	}

	// Here we know we have to execute the trial. In case the trial failed we
	// simply return the error. In case the trial was successful we register it to
	// be sure we already executed it. That causes the trial to be ignored the
	// next time the transaction is being executed and the transaction's replay is
	// being executed, if any.
	{
		output, err := config.Trial(ctx)
		if err != nil {
			return microerror.Mask(err)
		}
		e.logger.Log("debug", fmt.Sprintf("executed transaction trial for transaction ID '%s' and trial ID '%s'", transactionID, config.TrialID))

		rKey := transactionKey("transaction", transactionID, "trial", config.TrialID, "result")
		b, err := config.TrialEncoder(output)
		if err != nil {
			return microerror.Mask(err)
		}
		if b != nil {
			rVal := string(b)
			kv, err := microstorage.NewKV(rKey, rVal)
			if err != nil {
				return microerror.Mask(err)
			}
			err = e.storage.Put(ctx, kv)
			if err != nil {
				return microerror.Mask(err)
			}
		}

		tKey := transactionKey("transaction", transactionID, "trial", config.TrialID)
		kv, err := microstorage.NewKV(tKey, "{}")
		if err != nil {
			return microerror.Mask(err)
		}
		err = e.storage.Put(ctx, kv)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	return nil
}

func (e *executer) ExecuteConfig() ExecuteConfig {
	return ExecuteConfig{
		Replay:        nil,
		ReplayDecoder: DefaultReplayDecoder,
		Trial:         nil,
		TrialEncoder:  DefaultTrialEncoder,
		TrialID:       "",
	}
}

func validateExecuteConfig(config ExecuteConfig) error {
	if config.Replay != nil && config.ReplayDecoder == nil {
		return microerror.Maskf(invalidExecutionError, "replay decoder must not be empty when replay is given")
	}
	if config.Trial == nil {
		return microerror.Maskf(invalidExecutionError, "trial must not be empty")
	}
	if config.TrialEncoder == nil {
		return microerror.Maskf(invalidExecutionError, "trial encoder must not be empty")
	}
	if config.TrialID == "" {
		return microerror.Maskf(invalidExecutionError, "trial ID must not be empty")
	}

	return nil
}

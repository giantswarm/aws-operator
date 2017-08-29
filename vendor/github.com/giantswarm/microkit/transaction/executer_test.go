package transaction

import (
	"context"
	"reflect"
	"testing"

	transactionid "github.com/giantswarm/microkit/transaction/context/id"
	"github.com/giantswarm/micrologger/microloggertest"
	"github.com/giantswarm/microstorage/microstoragetest"
)

func Test_Executer_NoTransactionIDGiven(t *testing.T) {
	config := DefaultExecuterConfig()
	config.Logger = microloggertest.New()
	config.Storage = microstoragetest.New()
	newExecuter, err := NewExecuter(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	var replayExecuted int
	var trialExecuted int

	replay := func(context context.Context, v interface{}) error {
		replayExecuted++
		return nil
	}
	trial := func(context context.Context) (interface{}, error) {
		trialExecuted++
		return nil, nil
	}

	var ctx context.Context
	var executeConfig ExecuteConfig
	{
		ctx = context.Background()

		executeConfig = newExecuter.ExecuteConfig()
		executeConfig.Replay = replay
		executeConfig.Trial = trial
		executeConfig.TrialID = "test-trial-ID"
	}

	// The first execution of the transaction causes the trial to be executed
	// once. The replay function must not be executed at all.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}

	// There is no transaction ID provided, so the trial is executed again and the
	// replay function is still untouched.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 2 {
			t.Fatal("expected", 2, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}

	// There is no transaction ID provided, so the trial is executed again and the
	// replay function is still untouched.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 3 {
			t.Fatal("expected", 3, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}
}

func Test_Executer_TransactionIDGiven(t *testing.T) {
	config := DefaultExecuterConfig()
	config.Logger = microloggertest.New()
	config.Storage = microstoragetest.New()
	newExecuter, err := NewExecuter(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	var replayExecuted int
	var trialExecuted int

	replay := func(context context.Context, v interface{}) error {
		replayExecuted++
		return nil
	}
	trial := func(context context.Context) (interface{}, error) {
		trialExecuted++
		return nil, nil
	}

	var ctx context.Context
	var executeConfig ExecuteConfig
	{
		ctx = context.Background()
		ctx = transactionid.NewContext(ctx, "test-transaction-id")

		executeConfig = newExecuter.ExecuteConfig()
		executeConfig.Replay = replay
		executeConfig.Trial = trial
		executeConfig.TrialID = "test-trial-ID"
	}

	// The first execution of the transaction causes the trial to be executed
	// once. The replay function must not be executed at all.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 0 {
			t.Fatal("expected", 0, "got", replayExecuted)
		}
	}

	// There is a transaction ID provided, so the trial is not executed again and
	// the replay function is executed the first time.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 1 {
			t.Fatal("expected", 1, "got", replayExecuted)
		}
	}

	// There is a transaction ID provided, so the trial is still not executed
	// again and the replay function is executed the second time.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
		if replayExecuted != 2 {
			t.Fatal("expected", 2, "got", replayExecuted)
		}
	}
}

func Test_Executer_TransactionIDGiven_NoReplay(t *testing.T) {
	config := DefaultExecuterConfig()
	config.Logger = microloggertest.New()
	config.Storage = microstoragetest.New()
	newExecuter, err := NewExecuter(config)
	if err != nil {
		t.Fatal("expected", nil, "got", err)
	}

	var trialExecuted int

	trial := func(context context.Context) (interface{}, error) {
		trialExecuted++
		return nil, nil
	}

	var ctx context.Context
	var executeConfig ExecuteConfig
	{
		ctx = context.Background()
		ctx = transactionid.NewContext(ctx, "test-transaction-id")

		executeConfig = newExecuter.ExecuteConfig()
		executeConfig.Trial = trial
		executeConfig.TrialID = "test-trial-ID"
	}

	// The first execution of the transaction causes the trial to be executed
	// once. There is no replay function.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
	}

	// There is a transaction ID provided, so the trial is not executed again.
	// There is no replay function.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
	}

	// There is a transaction ID provided, so the trial is still not executed
	// again. There is no replay function.
	{
		err := newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		if trialExecuted != 1 {
			t.Fatal("expected", 1, "got", trialExecuted)
		}
	}
}

func Test_Executer_TransactionResult(t *testing.T) {
	tests := []struct {
		TrialOutput     interface{}
		WantReplayInput interface{}
	}{
		{ // 0
			TrialOutput:     []byte("hello world"),
			WantReplayInput: "hello world",
		},
		{ // 1
			TrialOutput:     float64(4.3),
			WantReplayInput: "4.3",
		},
		{ // 2
			TrialOutput:     nil,
			WantReplayInput: nil,
		},
		{ // 3
			TrialOutput:     "hello world",
			WantReplayInput: "hello world",
		},
		{ // 4
			TrialOutput:     "",
			WantReplayInput: "",
		},
		{ // 5
			TrialOutput: struct {
				Foo string `json:"foo"`
				Bar int    `json:"bar"`
			}{
				Foo: "foo-val",
				Bar: 43,
			},
			WantReplayInput: `{"foo":"foo-val","bar":43}`,
		},
	}

	for i, tc := range tests {
		config := DefaultExecuterConfig()
		config.Logger = microloggertest.New()
		config.Storage = microstoragetest.New()
		newExecuter, err := NewExecuter(config)
		if err != nil {
			t.Fatal("expected", nil, "got", err)
		}

		var replayInput interface{}

		replay := func(context context.Context, v interface{}) error {
			replayInput = v
			return nil
		}
		trial := func(context context.Context) (interface{}, error) {
			return tc.TrialOutput, nil
		}

		var ctx context.Context
		var executeConfig ExecuteConfig
		{
			ctx = context.Background()
			ctx = transactionid.NewContext(ctx, "test-transaction-id")

			executeConfig = newExecuter.ExecuteConfig()
			executeConfig.Replay = replay
			executeConfig.Trial = trial
			executeConfig.TrialID = "test-trial-ID"
		}

		err = newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Errorf("#%d: want nil, got %#v", i, err)
			continue
		}
		if replayInput != nil {
			t.Errorf("#%d: want %#v, got %#v", i, nil, replayInput)
			continue
		}

		err = newExecuter.Execute(ctx, executeConfig)
		if err != nil {
			t.Errorf("#%d: want nil, got %#v", i, err)
			continue
		}
		if !reflect.DeepEqual(tc.WantReplayInput, replayInput) {
			t.Errorf("#%d: want %#v, got %#v", i, tc.TrialOutput, replayInput)
			continue
		}
	}
}

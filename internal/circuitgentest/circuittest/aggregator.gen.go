// Code generated by circuitgen tool. DO NOT EDIT

package circuittest

import (
	"context"

	"github.com/cep21/circuit"
	"github.com/twitchtv/circuitgen/internal/circuitgentest"
)

// CircuitWrapperAggregatorConfig contains configuration for CircuitWrapperAggregator. All fields are optional
type CircuitWrapperAggregatorConfig struct {
	// ShouldSkipError determines whether an error should be skipped and have the circuit
	// track the call as successful. This takes precedence over IsBadRequest
	ShouldSkipError func(error) bool

	// IsBadRequest is an optional bad request checker. It is useful to not count user errors as faults
	IsBadRequest func(error) bool

	// Prefix is prepended to all circuit names
	Prefix string

	// Defaults are used for all created circuits. Per-circuit configs override this
	Defaults circuit.Config

	// CircuitIncSum is the configuration used for the IncSum circuit. This overrides values set by Defaults
	CircuitIncSum circuit.Config
}

// CircuitWrapperAggregator is a circuit wrapper for *circuitgentest.Aggregator
type CircuitWrapperAggregator struct {
	*circuitgentest.Aggregator

	// ShouldSkipError determines whether an error should be skipped and have the circuit
	// track the call as successful. This takes precedence over IsBadRequest
	ShouldSkipError func(error) bool

	// IsBadRequest checks whether to count a user error against the circuit. It is recommended to set this
	IsBadRequest func(error) bool

	// CircuitIncSum is the circuit for method IncSum
	CircuitIncSum *circuit.Circuit
}

// NewCircuitWrapperAggregator creates a new circuit wrapper and initializes circuits
func NewCircuitWrapperAggregator(
	manager *circuit.Manager,
	embedded *circuitgentest.Aggregator,
	conf CircuitWrapperAggregatorConfig,
) (*CircuitWrapperAggregator, error) {
	if conf.ShouldSkipError == nil {
		conf.ShouldSkipError = func(err error) bool {
			return false
		}
	}

	if conf.IsBadRequest == nil {
		conf.IsBadRequest = func(err error) bool {
			return false
		}
	}

	w := &CircuitWrapperAggregator{
		Aggregator:      embedded,
		ShouldSkipError: conf.ShouldSkipError,
		IsBadRequest:    conf.IsBadRequest,
	}

	var err error
	w.CircuitIncSum, err = manager.CreateCircuit(conf.Prefix+"Aggregator.IncSum", conf.CircuitIncSum, conf.Defaults)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// IncSum calls the embedded *circuitgentest.Aggregator's method IncSum with CircuitIncSum
func (w *CircuitWrapperAggregator) IncSum(ctx context.Context, p1 int) error {
	var skippedErr error

	err := w.CircuitIncSum.Run(ctx, func(ctx context.Context) error {
		err := w.Aggregator.IncSum(ctx, p1)

		if w.ShouldSkipError(err) {
			skippedErr = err
			return nil
		}

		if w.IsBadRequest(err) {
			return &circuit.SimpleBadRequest{Err: err}
		}
		return err
	})

	if skippedErr != nil {
		err = skippedErr
	}

	if berr, ok := err.(*circuit.SimpleBadRequest); ok {
		err = berr.Err
	}

	return err
}

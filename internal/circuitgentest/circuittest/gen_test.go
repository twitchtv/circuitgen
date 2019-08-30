// Copyright 2019 Twitch Interactive, Inc.  All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package circuittest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cep21/circuit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/twitchtv/circuitgen/internal/circuitgentest"
	"github.com/twitchtv/circuitgen/internal/circuitgentest/model"
	"github.com/twitchtv/circuitgen/internal/circuitgentest/rep"
)

// Test generated clients

func TestPublisherInterface(t *testing.T) {
	manager := &circuit.Manager{}

	ctx := context.Background()

	grants := map[circuitgentest.Seed][][]circuitgentest.Grant{
		"seed": {
			{
				{Name: "something"},
			},
		},
	}
	topics := circuitgentest.TopicsList{List: []string{"1234"}}
	publishInput := rep.PublishInput{UserID: "9999"}
	publishResult := &model.Result{Nonce: "abcdefg"}
	m := &circuitgentest.MockPublisher{}
	m.On("PublishWithResult", mock.Anything, publishInput).Return(publishResult, nil).Once()

	opt := rep.PublishOption{Sample: 0.1}
	opt2 := rep.PublishOption{Sample: 0.2}
	m.On("Publish", mock.Anything, grants, topics, opt, opt2).Return(nil, nil).Once()

	m.On("Close", mock.Anything).Return(nil).Once()

	publishCounter := &runMetricsCounter{}
	publishWithResultCounter := &runMetricsCounter{}
	publisher, err := NewCircuitWrapperPublisher(manager, m, CircuitWrapperPublisherConfig{
		Defaults: circuit.Config{
			Execution: circuit.ExecutionConfig{
				Timeout: 1 * time.Second,
			},
		},
		CircuitPublish: circuit.Config{
			Execution: circuit.ExecutionConfig{
				Timeout: 2 * time.Second,
			},
			Metrics: circuit.MetricsCollectors{
				Run: []circuit.RunMetrics{publishCounter},
			},
		},
		CircuitPublishWithResult: circuit.Config{
			Metrics: circuit.MetricsCollectors{
				Run: []circuit.RunMetrics{publishWithResultCounter},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, publisher)

	// Check circuit names
	names := circuitNames(manager)
	require.Contains(t, names, "Publisher.PublishWithResult")
	require.Contains(t, names, "Publisher.Publish")

	require.Equal(t, 1*time.Second, manager.GetCircuit("Publisher.PublishWithResult").Config().Execution.Timeout)
	require.Equal(t, 2*time.Second, manager.GetCircuit("Publisher.Publish").Config().Execution.Timeout)

	_, err = publisher.Publish(ctx, grants, topics, opt, opt2)
	require.NoError(t, err)

	res, err := publisher.PublishWithResult(ctx, publishInput)
	require.NoError(t, err)
	require.Equal(t, publishResult, res)

	require.NoError(t, publisher.Close())

	// Check embedded called
	m.AssertExpectations(t)

	// Check circuits called
	assert.EqualValues(t, 1, publishCounter.success)
	assert.EqualValues(t, 1, publishWithResultCounter.success)
}

func TestPublisherInterfaceErrors(t *testing.T) {
	manager := &circuit.Manager{}

	testError := errors.New("test error")
	m := &circuitgentest.MockPublisher{}
	m.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil, testError).Once()

	publishCounter := &runMetricsCounter{}
	publisher, err := NewCircuitWrapperPublisher(manager, m, CircuitWrapperPublisherConfig{
		CircuitPublish: circuit.Config{
			Metrics: circuit.MetricsCollectors{
				Run: []circuit.RunMetrics{publishCounter},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, publisher)

	ctx := context.Background()
	_, err = publisher.Publish(ctx, map[circuitgentest.Seed][][]circuitgentest.Grant{}, circuitgentest.TopicsList{})
	require.Equal(t, testError, err)

	// Check embedded called
	m.AssertExpectations(t)

	// Check circuit called
	assert.EqualValues(t, 1, publishCounter.failure)
}

func TestPublisherInterfaceBadRequest(t *testing.T) {
	manager := &circuit.Manager{}

	testError := errors.New("bad request error")
	m := &circuitgentest.MockPublisher{}
	m.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, testError).Once()

	publishCounter := &runMetricsCounter{}
	publisher, err := NewCircuitWrapperPublisher(manager, m, CircuitWrapperPublisherConfig{
		IsBadRequest: func(err error) bool {
			return err == testError
		},
		CircuitPublish: circuit.Config{
			Metrics: circuit.MetricsCollectors{
				Run: []circuit.RunMetrics{publishCounter},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, publisher)

	ctx := context.Background()
	_, err = publisher.Publish(ctx, map[circuitgentest.Seed][][]circuitgentest.Grant{}, circuitgentest.TopicsList{})
	require.Equal(t, testError, err)

	// Check embedded called
	m.AssertExpectations(t)

	// Check circuit called
	assert.EqualValues(t, 1, publishCounter.badRequest)
}

// Thinner test of aliased wrapper.
func TestPubsubInterface(t *testing.T) {
	manager := &circuit.Manager{}

	m := &circuitgentest.MockPublisher{}
	m.On("PublishWithResult", mock.Anything, mock.Anything).Return(nil, nil).Once()
	m.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
	m.On("Close", mock.Anything).Return(nil).Once()

	pubsub, err := NewCircuitWrapperPubsub(manager, m, CircuitWrapperPubsubConfig{})
	require.NoError(t, err)
	require.NotNil(t, pubsub)

	// Check circuit names
	names := circuitNames(manager)
	require.Contains(t, names, "Pubsub.PublishWithResult")
	require.Contains(t, names, "Pubsub.Publish")

	ctx := context.Background()
	_, err = pubsub.Publish(ctx, map[circuitgentest.Seed][][]circuitgentest.Grant{}, circuitgentest.TopicsList{})
	require.NoError(t, err)

	_, err = pubsub.PublishWithResult(ctx, rep.PublishInput{})
	require.NoError(t, err)

	require.NoError(t, pubsub.Close())

	// Check embedded called
	m.AssertExpectations(t)
}

func TestAggregatorStruct(t *testing.T) {
	manager := &circuit.Manager{}
	agg := &circuitgentest.Aggregator{}

	incSumCounter := &runMetricsCounter{}
	wrapperAgg, err := NewCircuitWrapperAggregator(manager, agg, CircuitWrapperAggregatorConfig{
		CircuitIncSum: circuit.Config{
			Metrics: circuit.MetricsCollectors{
				Run: []circuit.RunMetrics{incSumCounter},
			},
		},
	})
	require.NoError(t, err)

	err = wrapperAgg.IncSum(context.Background(), 10)
	require.NoError(t, err)
	require.Equal(t, 10, agg.Sum())
	require.Equal(t, 10, wrapperAgg.Sum())
	require.Equal(t, 1, incSumCounter.success)

	sumErr := errors.New("sum error")
	agg.IncSumError = sumErr
	err = wrapperAgg.IncSum(context.Background(), 10)
	require.Equal(t, sumErr, err)

}

func circuitNames(m *circuit.Manager) []string {
	names := make([]string, 0, len(m.AllCircuits()))
	for _, circ := range m.AllCircuits() {
		names = append(names, circ.Name())
	}
	return names
}

type runMetricsCounter struct {
	success                int
	failure                int
	timeout                int
	badRequest             int
	interrupt              int
	concurrencyLimitReject int
	shortCircuit           int
}

func (r *runMetricsCounter) Success(now time.Time, duration time.Duration)       { r.success++ }
func (r *runMetricsCounter) ErrFailure(now time.Time, duration time.Duration)    { r.failure++ }
func (r *runMetricsCounter) ErrTimeout(now time.Time, duration time.Duration)    { r.timeout++ }
func (r *runMetricsCounter) ErrBadRequest(now time.Time, duration time.Duration) { r.badRequest++ }
func (r *runMetricsCounter) ErrInterrupt(now time.Time, duration time.Duration)  { r.interrupt++ }
func (r *runMetricsCounter) ErrConcurrencyLimitReject(now time.Time)             { r.concurrencyLimitReject++ }
func (r *runMetricsCounter) ErrShortCircuit(now time.Time)                       { r.shortCircuit++ }

var _ circuit.RunMetrics = (*runMetricsCounter)(nil)

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

package circuittestmulti

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cep21/circuit"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/twitchtv/circuitgen/internal/circuitgentest"
	"github.com/twitchtv/circuitgen/internal/circuitgentest/rep"
)

// Thinner test of multi-gen interface
func TestPublisherInterface(t *testing.T) {
	manager := &circuit.Manager{}

	m := &circuitgentest.MockPublisher{}
	m.On("PublishWithResult", mock.Anything, mock.Anything).Return(nil, nil).Once()
	m.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
	m.On("Close", mock.Anything).Return(nil).Once()

	publisher, err := NewCircuitWrapperPublisher(manager, m, CircuitWrapperPublisherConfig{})
	require.NoError(t, err)
	require.NotNil(t, publisher)

	// Check circuit names
	names := circuitNames(manager)
	require.Contains(t, names, "Publisher.PublishWithResult")
	require.Contains(t, names, "Publisher.Publish")

	ctx := context.Background()
	_, err = publisher.Publish(ctx, map[circuitgentest.Seed][][]circuitgentest.Grant{}, circuitgentest.TopicsList{})
	require.NoError(t, err)

	_, err = publisher.PublishWithResult(ctx, rep.PublishInput{})
	require.NoError(t, err)

	require.NoError(t, publisher.Close())

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

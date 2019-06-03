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

package circuitgentest

import (
	"context"

	"github.com/twitchtv/circuitgen/internal/circuitgentest/model"
	"github.com/twitchtv/circuitgen/internal/circuitgentest/rep"
	"github.com/stretchr/testify/mock"
)

// TopicsList is a test struct
type TopicsList struct {
	List []string
}

// Seed is a test type
type Seed string

// Grant is a test struct
type Grant struct {
	// Name is a name
	Name string
}

// Publisher is an interface for testing. This interface has many different types to test generation.
type Publisher interface {
	// PublishWithResult is a test method and should be wrapped
	PublishWithResult(context.Context, rep.PublishInput) (*model.Result, error)
	// PublishWithResult is a test method and should be wrapped
	Publish(context.Context, map[Seed][][]Grant, TopicsList, ...rep.PublishOption) error
	// PublishWithResult is a test method and should not be wrapped
	Close() error
}

// MockPublisher is a test mock for the Publisher interface
type MockPublisher struct {
	mock.Mock
}

// PublishWithResult mocks the method
func (m *MockPublisher) PublishWithResult(ctx context.Context, input rep.PublishInput) (*model.Result, error) {
	args := m.Called(ctx, input)
	var r0 *model.Result
	if args.Get(0) != nil {
		var ok bool
		r0, ok = args.Get(0).(*model.Result)
		if !ok {
			panic("args.Get(0) is not a *model.Result")
		}
	}
	return r0, args.Error(1)
}

// Publish mocks the method
func (m *MockPublisher) Publish(ctx context.Context, g map[Seed][][]Grant, s TopicsList, opts ...rep.PublishOption) error {
	var ca []interface{}
	ca = append(ca, ctx, g, s)
	va := make([]interface{}, len(opts))
	for i := range opts {
		va[i] = opts[i]
	}
	ca = append(ca, va...)
	args := m.Called(ca...)
	return args.Error(0)
}

// Close mocks the method
func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

var _ Publisher = (*MockPublisher)(nil)

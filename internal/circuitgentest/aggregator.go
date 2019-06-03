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
)

// Aggregator is a test struct for wrapper generation
type Aggregator struct {
	IncSumError error
	sum         int
}

// IncSum increments sum by v
func (a *Aggregator) IncSum(ctx context.Context, v int) error {
	a.sum += v
	return a.IncSumError
}

// Sum returns sum
func (a *Aggregator) Sum() int {
	return a.sum
}

func (a *Aggregator) privateIncSum(_ context.Context, v int) error {
	a.sum += v
	return a.IncSumError
}

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

// Test generation in the same package. There are only compile-time tests. Comprehensive tests for
// circuit wrappers are in circuitest/gen_test.go
//
// Disable goimports to catch any import bugs

//go:generate circuitgen circuit --goimports=true --pkg . --name Publisher --out ./publisher.gen.go
//go:generate circuitgen circuit --goimports=true --pkg . --name Aggregator --out ./

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

package main

import (
	"fmt"
	"strings"
)

// This file contains structs used in the wrapper generating templates.

// TypeMetadata stores metadata about a Go type.
type TypeMetadata struct {
	// Name of the package this type is defined in.
	PackageName string

	// Path to the package this type is defined in.
	PackagePath string

	// Holds type information about this type
	TypeInfo TypeInfo

	// Imports of this type from all the methods
	Imports []Import

	// Methods of this type
	Methods []Method
}

// Method represents a method function on a type
type Method struct {
	// The name of the method on the type
	Name string

	// Input params
	Params []TypeInfo

	// Return results
	Results []TypeInfo

	// Whether this method is variadic
	Variadic bool
}

// TypeInfo stores the name and whether it is an interface
type TypeInfo struct {
	// The name of the type with or without the package qualifier. The qualifier is set appropriately
	// Examples:
	//
	//		"aws.Context" or "Context"
	//		"error would be "error"
	Name string

	// The name of the type without the package qualifier. Ex. "aws.Context" would be "Context"
	NameWithoutQualifier string

	// Whether this type is an interface
	IsInterface bool
}

// Import represents a package import
type Import struct {
	// ex. "github.com/aws/aws-sdk-go/service/dynamodb"
	Path string
}

// ParamsSignature generates the signature for the methods params
// ex. "r0 aws.Context, r1 *dynamodb.BatchGetItemInput"
func (m Method) ParamsSignature(overrides ...string) string {
	s := ""
	mt := m.Params
	l := len(mt)

	for i := 0; i < l; i++ {
		varName := fmt.Sprintf("p%d", i)

		if i < len(overrides) {
			varName = overrides[i]
		}

		if i == l-1 && m.Variadic {
			const sliceChars = "[]"
			// the qualfied name contains the "[]" at the beginning, so chop it off
			s += fmt.Sprintf("%s ...%s", varName, mt[i].Name[len(sliceChars):])
		} else {
			s += fmt.Sprintf("%s %s", varName, mt[i].Name)
			if i < l-1 {
				s += ", "
			}
		}
	}

	return s
}

// CallSignatureWithClosure generates the signature for calling the embedded interface with a closure
func (m Method) CallSignatureWithClosure() string {
	s := ""
	mt := m.Params
	l := len(mt)

	for i := 0; i < l; i++ {
		if i == l-1 && m.Variadic {
			s += fmt.Sprintf("p%d...", i)
		} else {
			if i == 0 {
				s += "ctx"
			} else {
				s += fmt.Sprintf("p%d", i)
			}

			if i < l-1 {
				s += ", "
			}
		}
	}

	return s
}

// ResultsSignature generates the signature for the methods results
// ex. "(*dynamodb.BatchGetItemOutput, error)"
func (m Method) ResultsSignature() string {
	mt := m.Results
	if len(mt) == 1 {
		return mt[0].Name
	}

	s := "("
	l := len(mt)
	for i := 0; i < l; i++ {
		if i == l-1 {
			s += mt[i].Name
		} else {
			s += mt[i].Name + ", "
		}
	}

	return s + ")"
}

// ResultsClosureVariableDeclarations generates the variable declarations needed for making a call with a closure. These variables are needed
// outside the function scope (ex. (*circuit.Circuit).Run()).
// ex. "var r0 *dynamodb.UpdateItemInput
func (m Method) ResultsClosureVariableDeclarations() string {
	s := ""
	for i, t := range m.Results[:len(m.Results)-1] {
		s += fmt.Sprintf("var r%d %s\n", i, t.Name)
	}

	return s
}

// HasOneMethodResultVariable returns whether there is exactly one return value
func (m Method) HasOneMethodResultVariable() bool {
	return len(m.Results) == 1
}

// ResultsCircuitVariableAssignments generates the variable names needed when assigning the embedded interface method call.
// ex. "r0, err"
func (m Method) ResultsCircuitVariableAssignments() string {
	s := ""

	for i := range m.Results[:len(m.Results)-1] {
		s += fmt.Sprintf("r%d, ", i)
	}

	s += "err"

	return s
}

// ResultsClosureVariableReturns generates the variable names needed when returning the results of a closure wrapped method call.
// ex. "r0, err"
func (m Method) ResultsClosureVariableReturns() string {
	s := ""
	for i := range m.Results[:len(m.Results)-1] {
		s += fmt.Sprintf("r%d, ", i)
	}

	return s
}

// IsWrappingSupported returns true only if the method supports context and returns an error.
func (m Method) IsWrappingSupported() bool {
	if len(m.Params) == 0 || len(m.Results) == 0 {
		return false
	}

	supportsContext := strings.HasSuffix(m.Params[0].Name, "Context")
	returnsAnError := m.Results[len(m.Results)-1].Name == "error"

	return supportsContext && returnsAnError
}

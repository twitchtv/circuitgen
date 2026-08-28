// Copyright 2026 Twitch Interactive, Inc.  All Rights Reserved.
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
	"go/token"
	"go/types"
	"reflect"
	"testing"
)

func TestResolvePkgPathsAlias(t *testing.T) {
	targetPkg := types.NewPackage("example.com/target", "target")
	target := types.NewNamed(
		types.NewTypeName(token.NoPos, targetPkg, "Target", nil),
		types.Typ[types.String],
		nil,
	)
	aliasPkg := types.NewPackage("example.com/alias", "alias")
	alias := types.NewAlias(
		types.NewTypeName(token.NoPos, aliasPkg, "Alias", nil),
		target,
	)

	paths, err := resolvePkgPaths(alias)
	if err != nil {
		t.Fatalf("resolvePkgPaths returned an error for an alias: %v", err)
	}
	if want := []string{"example.com/alias"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("resolvePkgPaths returned %v, want %v", paths, want)
	}
	if got, want := typeInfo(alias, "example.com/output").Name, "alias.Alias"; got != want {
		t.Fatalf("typeInfo returned %q, want %q", got, want)
	}
}

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
	"errors"
	"fmt"
	"go/types"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages" // latest loader that supports modules
	"golang.org/x/tools/go/types/typeutil"
)

func resolvePackagePath(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	conf := &packages.Config{
		Mode: packages.NeedName,
	}

	if strings.HasSuffix(path, ".go") {
		path = filepath.Dir(path)
	}

	pkgs, err := packages.Load(conf, path)
	if err != nil {
		return "", err
	}

	// We don't check pkgs[0].Errors because the package may not exist
	return pkgs[0].PkgPath, nil
}

func loadPackages(pkgPaths ...string) ([]*packages.Package, error) {
	conf := &packages.Config{
		Mode: packages.LoadTypes,
	}

	pkgs, err := packages.Load(conf, pkgPaths...)
	if err != nil {
		return nil, fmt.Errorf("loading packages: %v", err)
	}

	return pkgs, nil
}

func firstPackagesError(pkgs []*packages.Package) error {
	var err error

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, pkgErr := range pkg.Errors {
			if err == nil {
				err = fmt.Errorf("loading package %s: %v", pkg.PkgPath, pkgErr)
			}
		}
	})

	return err
}

// Parse the type for its type info, imports, and methods.
func parseType(t types.Type, outPkgPath string) (TypeMetadata, error) {
	mset := methodSet(t)
	if len(mset) == 0 {
		return TypeMetadata{}, fmt.Errorf("empty methodset. %v has no exported methods", t)
	}

	// Get all the import paths
	imports, err := parseImports(t, mset, outPkgPath)
	if err != nil {
		return TypeMetadata{}, err
	}

	methods := make([]Method, 0, len(mset))

	// For each method of the interface, get the name, params, results, and
	// other info needed to generate a wrapper.
	for _, m := range mset {
		sig, ok := m.Type().(*types.Signature)
		if !ok {
			return TypeMetadata{}, fmt.Errorf("method %s is not a signature", m.String())
		}

		methods = append(methods, Method{
			Name:     m.Obj().Name(),
			Params:   parseTuple(sig.Params(), outPkgPath),
			Results:  parseTuple(sig.Results(), outPkgPath),
			Variadic: sig.Variadic(),
		})
	}

	tn, ok := t.(*types.Named)
	if !ok {
		return TypeMetadata{}, errors.New("not a named type")
	}

	tm := TypeMetadata{
		PackageName: tn.Obj().Pkg().Name(),
		PackagePath: stripVendor(tn.Obj().Pkg().Path()),
		TypeInfo:    typeInfo(t, outPkgPath),
		Imports:     imports,
		Methods:     methods,
	}

	return tm, nil
}

func methodSet(t types.Type) []*types.Selection {
	var mset []*types.Selection
	if types.IsInterface(t) {
		mset = typeutil.IntuitiveMethodSet(t.Underlying(), nil)
	} else {
		mset = typeutil.IntuitiveMethodSet(t, nil) // supports structs
	}

	var exported []*types.Selection
	for _, m := range mset {
		if m.Obj().Exported() {
			exported = append(exported, m)
		}
	}

	return exported
}

// get all the package import paths given the type.
func resolvePkgPaths(p types.Type) ([]string, error) {
	switch t := p.(type) {
	case *types.Signature:
		var r []string
		for i := 0; i < t.Params().Len(); i++ {
			paths, err := resolvePkgPaths(t.Params().At(i).Type())
			if err != nil {
				return nil, err
			}
			r = append(r, paths...)
		}
		for i := 0; i < t.Results().Len(); i++ {
			paths, err := resolvePkgPaths(t.Results().At(i).Type())
			if err != nil {
				return nil, err
			}
			r = append(r, paths...)
		}
		return r, nil
	case *types.Pointer:
		return resolvePkgPaths(t.Elem())
	case *types.Map:
		keyPaths, err := resolvePkgPaths(t.Key())
		if err != nil {
			return nil, err
		}
		elemPaths, err := resolvePkgPaths(t.Elem())
		if err != nil {
			return nil, err
		}
		return append(keyPaths, elemPaths...), nil
	case *types.Slice:
		return resolvePkgPaths(t.Elem())
	case *types.Named:
		// builtins (e.g. error) have a nil package
		if pkg := t.Obj().Pkg(); pkg != nil {
			return []string{stripVendor(pkg.Path())}, nil
		}
	case *types.Basic:
	case *types.Interface:
	case *types.Struct: // struct{}
		// Break out of the switch and return below
	default:
		return nil, fmt.Errorf("resolvePkgPaths: invalid type: %v", t)
	}

	return []string{}, nil
}

// returns unique import package paths for the given type
func parseImports(t types.Type, mset []*types.Selection, outPkgPath string) ([]Import, error) {
	pkgPaths := make([]string, 0, len(mset))
	for _, m := range mset {
		paths, err := resolvePkgPaths(m.Type())
		if err != nil {
			return nil, err
		}
		pkgPaths = append(pkgPaths, paths...)
	}

	var imports []Import
	inPkg := typePackagePath(t) == outPkgPath
	if !inPkg {
		// add import for the type
		imports = append(imports, Import{
			Path: typePackagePath(t),
		})
	}

	for _, path := range uniqueStringSlice(pkgPaths) {
		// Don't add import if in the same package to prevent a circular dependency
		if !inPkg || (inPkg && path != outPkgPath) {
			imports = append(imports, Import{Path: path})
		}
	}

	return imports, nil
}

// parseTuple parses a list of variables, like a method's params and results.
func parseTuple(tuple *types.Tuple, outPkgPath string) []TypeInfo {
	vars := []TypeInfo{}

	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)

		vars = append(vars, typeInfo(v.Type(), outPkgPath))
	}

	return vars
}

func typeInfo(t types.Type, outPkgPath string) TypeInfo {
	return TypeInfo{
		// Check if the pkg qualifier should be added depending on
		// if the type is in outPkgPath
		Name: types.TypeString(t, func(pkg *types.Package) string {
			if pkg.Path() == outPkgPath {
				return ""
			}
			return pkg.Name()
		}),
		NameWithoutQualifier: types.TypeString(t, func(pkg *types.Package) string {
			return ""
		}),
		IsInterface: types.IsInterface(t),
	}
}

func uniqueStringSlice(s []string) []string {
	m := map[string]struct{}{}
	paths := []string{}
	for _, p := range s {
		if _, ok := m[p]; !ok {
			paths = append(paths, p)
			m[p] = struct{}{}
		}
	}

	return paths
}

// Returns the package path of the given type. The type must be named
func typePackagePath(t types.Type) string {
	tn, ok := t.(*types.Named)
	if !ok {
		return ""
	}
	if tn.Obj() == nil {
		return ""
	}
	if tn.Obj().Pkg() == nil {
		return ""
	}
	return stripVendor(tn.Obj().Pkg().Path())
}

// stripVendor resolves the given package path by stripping vendor prefixes if needed.
func stripVendor(path string) string {
	const vendor = "/vendor/"
	i := strings.Index(path, vendor)
	if i > -1 {
		return path[i+len(vendor):]
	}

	return path
}

func circuitVersionSuffix(majorVersion int) string {
	if majorVersion < 3 {
		return ""
	}
	return "/v" + strconv.Itoa(majorVersion)
}

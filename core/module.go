// Package core provides the core Copybara module with transformations.
//
// The core module provides the fundamental transformation functions:
//   - core.workflow() - Define a migration workflow
//   - core.move() - Move/rename files
//   - core.copy() - Copy files
//   - core.replace() - Search and replace in files
//   - core.remove() - Remove files
//   - core.verify_match() - Verify regex matches in files
//   - core.transform() - Apply transformations
//   - core.reverse() - Reverse a transformation
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/CoreModule.java
package core

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the core.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "core",
	Members: starlark.StringDict{
		"workflow":     starlark.NewBuiltin("core.workflow", workflowFn),
		"move":         starlark.NewBuiltin("core.move", moveFn),
		"copy":         starlark.NewBuiltin("core.copy", copyFn),
		"replace":      starlark.NewBuiltin("core.replace", replaceFn),
		"remove":       starlark.NewBuiltin("core.remove", removeFn),
		"verify_match": starlark.NewBuiltin("core.verify_match", verifyMatchFn),
		"glob":         starlark.NewBuiltin("core.glob", globFn),
	},
}

// Globals returns global functions that should be available at the top level.
// This includes glob() which is commonly used as a global function.
func Globals() starlark.StringDict {
	return starlark.StringDict{
		"glob": starlark.NewBuiltin("glob", globFn),
	}
}

// moveFn implements core.move().
//
// Reference: CopyOrMove.java
func moveFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		before    string
		after     string
		paths     starlark.Value = starlark.None
		overwrite bool
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"before", &before,
		"after", &after,
		"paths?", &paths,
		"overwrite?", &overwrite,
	); err != nil {
		return nil, err
	}

	if before == after {
		return nil, fmt.Errorf("moving from the same folder to the same folder is a noop")
	}

	move := &Move{
		before:    before,
		after:     after,
		overwrite: overwrite,
	}

	// Handle paths parameter
	switch v := paths.(type) {
	case starlark.NoneType:
		move.paths = AllFiles()
	case *Glob:
		move.paths = v
	case *starlark.List:
		patterns := make([]string, v.Len())
		for i := range v.Len() {
			s, ok := starlark.AsString(v.Index(i))
			if !ok {
				return nil, fmt.Errorf("paths must be strings, got %s", v.Index(i).Type())
			}
			patterns[i] = s
		}
		var err error
		move.paths, err = NewGlob(patterns, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("paths must be a glob or list of strings, got %s", paths.Type())
	}

	return move, nil
}

// replaceFn implements core.replace().
//
// Reference: Replace.java
func replaceFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		before string
		after  string
		paths  starlark.Value = starlark.None
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"before", &before,
		"after", &after,
		"paths?", &paths,
	); err != nil {
		return nil, err
	}

	replace := &Replace{
		before: before,
		after:  after,
	}

	// Handle paths parameter
	switch v := paths.(type) {
	case starlark.NoneType:
		replace.paths = AllFiles()
	case *Glob:
		replace.paths = v
	case *starlark.List:
		patterns := make([]string, v.Len())
		for i := range v.Len() {
			s, ok := starlark.AsString(v.Index(i))
			if !ok {
				return nil, fmt.Errorf("paths must be strings, got %s", v.Index(i).Type())
			}
			patterns[i] = s
		}
		var err error
		replace.paths, err = NewGlob(patterns, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("paths must be a glob or list of strings, got %s", paths.Type())
	}

	return replace, nil
}

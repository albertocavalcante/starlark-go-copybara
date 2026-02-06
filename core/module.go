// Package core provides the core Copybara module with transformations.
//
// The core module provides the fundamental transformation functions:
//   - core.workflow() - Define a migration workflow
//   - core.move() - Move/rename files
//   - core.replace() - Search and replace in files
//   - core.transform() - Apply transformations
//   - core.reverse() - Reverse a transformation
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/CoreModule.java
package core

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the core.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "core",
	Members: starlark.StringDict{
		"workflow": starlark.NewBuiltin("core.workflow", workflowFn),
		"move":     starlark.NewBuiltin("core.move", moveFn),
		"replace":  starlark.NewBuiltin("core.replace", replaceFn),
	},
}

// workflowFn implements core.workflow().
//
// Reference: Workflow.java
func workflowFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		name            string
		origin          starlark.Value
		destination     starlark.Value
		authoring       starlark.Value
		transformations *starlark.List
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"origin?", &origin,
		"destination?", &destination,
		"authoring?", &authoring,
		"transformations?", &transformations,
	); err != nil {
		return nil, err
	}

	wf := &Workflow{
		name: name,
	}

	return wf, nil
}

// moveFn implements core.move().
//
// Reference: CopyOrMove.java
func moveFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var before, after string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"before", &before,
		"after", &after,
	); err != nil {
		return nil, err
	}

	return &Move{
		before: before,
		after:  after,
	}, nil
}

// replaceFn implements core.replace().
//
// Reference: Replace.java
func replaceFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var before, after string
	var paths *starlark.List

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"before", &before,
		"after", &after,
		"paths?", &paths,
	); err != nil {
		return nil, err
	}

	return &Replace{
		before: before,
		after:  after,
	}, nil
}

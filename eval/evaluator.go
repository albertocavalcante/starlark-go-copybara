// Package eval provides the Starlark evaluation engine for Copybara.
package eval

import (
	"go.starlark.net/starlark"
)

// Evaluator evaluates Starlark files.
type Evaluator struct {
	predeclared starlark.StringDict
	modules     map[string]starlark.StringDict
}

// New creates a new evaluator.
func New() *Evaluator {
	return &Evaluator{
		predeclared: make(starlark.StringDict),
		modules:     make(map[string]starlark.StringDict),
	}
}

// AddPredeclared adds a predeclared value.
func (e *Evaluator) AddPredeclared(name string, value starlark.Value) {
	e.predeclared[name] = value
}

// Eval evaluates a Starlark file.
func (e *Evaluator) Eval(filename string, src any) (starlark.StringDict, error) {
	thread := &starlark.Thread{
		Name: "copybara",
		Load: e.load,
	}

	return starlark.ExecFile(thread, filename, src, e.predeclared)
}

// load implements module loading.
func (e *Evaluator) load(thread *starlark.Thread, module string) (starlark.StringDict, error) {
	if m, ok := e.modules[module]; ok {
		return m, nil
	}

	// TODO: Implement file-based module loading
	return nil, nil
}

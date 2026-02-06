// Package copybara provides a Starlark interpreter for Copybara configuration files.
//
// This package is the main entry point for evaluating copy.bara.sky files
// and executing code transformation workflows.
//
// Reference: https://github.com/google/copybara
package copybara

import (
	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
)

// Interpreter evaluates Copybara configuration files.
type Interpreter struct {
	predeclared starlark.StringDict
}

// Result contains the evaluated configuration.
type Result struct {
	workflows []*core.Workflow
}

// New creates a new Copybara interpreter with default configuration.
func New(opts ...Option) *Interpreter {
	i := &Interpreter{
		predeclared: make(starlark.StringDict),
	}

	// Apply options
	for _, opt := range opts {
		opt(i)
	}

	// Register default modules
	i.registerDefaults()

	return i
}

// registerDefaults registers the default Copybara modules.
func (i *Interpreter) registerDefaults() {
	// TODO: Register core, git, metadata, authoring, folder modules
	i.predeclared["core"] = core.Module
}

// Eval evaluates a Copybara configuration file.
func (i *Interpreter) Eval(filename string, src any) (*Result, error) {
	thread := &starlark.Thread{
		Name: "copybara",
	}

	_, err := starlark.ExecFile(thread, filename, src, i.predeclared)
	if err != nil {
		return nil, err
	}

	// TODO: Extract workflows from globals

	return &Result{}, nil
}

// Workflows returns all workflows defined in the configuration.
func (r *Result) Workflows() []*core.Workflow {
	return r.workflows
}

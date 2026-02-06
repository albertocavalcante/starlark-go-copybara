// Package copybara provides a Starlark interpreter for Copybara configuration files.
//
// This package is the main entry point for evaluating copy.bara.sky files
// and executing code transformation workflows.
//
// Reference: https://github.com/google/copybara
package copybara

import (
	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/authoring"
	"github.com/albertocavalcante/starlark-go-copybara/core"
	"github.com/albertocavalcante/starlark-go-copybara/folder"
	"github.com/albertocavalcante/starlark-go-copybara/git"
	"github.com/albertocavalcante/starlark-go-copybara/metadata"
)

// Interpreter evaluates Copybara configuration files.
type Interpreter struct {
	predeclared starlark.StringDict
	dryRun      bool
	workDir     string
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
	i.predeclared["core"] = core.Module
	i.predeclared["git"] = git.Module
	i.predeclared["metadata"] = metadata.Module
	i.predeclared["authoring"] = authoring.Module
	i.predeclared["folder"] = folder.Module

	// Also register globals like glob()
	for name, val := range core.Globals() {
		i.predeclared[name] = val
	}
}

// Eval evaluates a Copybara configuration file.
func (i *Interpreter) Eval(filename string, src any) (*Result, error) {
	thread := &starlark.Thread{
		Name: "copybara",
	}

	globals, err := starlark.ExecFile(thread, filename, src, i.predeclared)
	if err != nil {
		return nil, err
	}

	// Extract workflows from globals
	var workflows []*core.Workflow
	for _, val := range globals {
		if wf, ok := val.(*core.Workflow); ok {
			workflows = append(workflows, wf)
		}
	}

	return &Result{workflows: workflows}, nil
}

// DryRun returns whether dry-run mode is enabled.
func (i *Interpreter) DryRun() bool {
	return i.dryRun
}

// WorkDir returns the working directory for file operations.
func (i *Interpreter) WorkDir() string {
	return i.workDir
}

// Workflows returns all workflows defined in the configuration.
func (r *Result) Workflows() []*core.Workflow {
	return r.workflows
}

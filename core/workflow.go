package core

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Workflow represents a Copybara migration workflow.
//
// A workflow defines how code is migrated from an origin to a destination,
// including the transformations to apply.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Workflow.java
type Workflow struct {
	name            string
	origin          starlark.Value
	destination     starlark.Value
	authoring       starlark.Value
	transformations []Transformation
	mode            WorkflowMode
}

// WorkflowMode defines how the workflow processes changes.
type WorkflowMode int

const (
	// ModeSquash squashes all changes into a single commit.
	ModeSquash WorkflowMode = iota
	// ModeIterative processes each change individually.
	ModeIterative
	// ModeChangeRequest creates change requests (PRs).
	ModeChangeRequest
)

// String implements starlark.Value.
func (w *Workflow) String() string {
	return fmt.Sprintf("workflow(%q)", w.name)
}

// Type implements starlark.Value.
func (w *Workflow) Type() string {
	return "workflow"
}

// Freeze implements starlark.Value.
func (w *Workflow) Freeze() {
	// Workflows are immutable after creation
}

// Truth implements starlark.Value.
func (w *Workflow) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (w *Workflow) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: workflow")
}

// Name returns the workflow name.
func (w *Workflow) Name() string {
	return w.name
}

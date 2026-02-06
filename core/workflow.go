package core

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// Workflow represents a Copybara migration workflow.
//
// A workflow defines how code is migrated from an origin to a destination,
// including the transformations to apply.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Workflow.java
type Workflow struct {
	name             string
	origin           starlark.Value
	destination      starlark.Value
	authoring        starlark.Value
	transformations  []Transformation
	originFiles      *Glob
	destinationFiles *Glob
	mode             WorkflowMode
	reversibleCheck  bool
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
	// ModeChangeRequestFromSOT imports pending changes from Source-of-Truth.
	ModeChangeRequestFromSOT
)

// String returns the string representation of the mode.
func (m WorkflowMode) String() string {
	switch m {
	case ModeSquash:
		return "SQUASH"
	case ModeIterative:
		return "ITERATIVE"
	case ModeChangeRequest:
		return "CHANGE_REQUEST"
	case ModeChangeRequestFromSOT:
		return "CHANGE_REQUEST_FROM_SOT"
	default:
		return "UNKNOWN"
	}
}

// ParseWorkflowMode parses a string to WorkflowMode.
func ParseWorkflowMode(s string) (WorkflowMode, error) {
	switch strings.ToUpper(s) {
	case "SQUASH":
		return ModeSquash, nil
	case "ITERATIVE":
		return ModeIterative, nil
	case "CHANGE_REQUEST":
		return ModeChangeRequest, nil
	case "CHANGE_REQUEST_FROM_SOT":
		return ModeChangeRequestFromSOT, nil
	default:
		return ModeSquash, fmt.Errorf("unknown workflow mode: %q", s)
	}
}

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
	if w.originFiles != nil {
		w.originFiles.Freeze()
	}
	if w.destinationFiles != nil {
		w.destinationFiles.Freeze()
	}
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

// Origin returns the origin value.
func (w *Workflow) Origin() starlark.Value {
	return w.origin
}

// Destination returns the destination value.
func (w *Workflow) Destination() starlark.Value {
	return w.destination
}

// Authoring returns the authoring value.
func (w *Workflow) Authoring() starlark.Value {
	return w.authoring
}

// Transformations returns the list of transformations.
func (w *Workflow) Transformations() []Transformation {
	return w.transformations
}

// OriginFiles returns the origin files glob.
func (w *Workflow) OriginFiles() *Glob {
	return w.originFiles
}

// DestinationFiles returns the destination files glob.
func (w *Workflow) DestinationFiles() *Glob {
	return w.destinationFiles
}

// Mode returns the workflow mode.
func (w *Workflow) Mode() WorkflowMode {
	return w.mode
}

// ReversibleCheck returns whether reversible check is enabled.
func (w *Workflow) ReversibleCheck() bool {
	return w.reversibleCheck
}

// workflowFn implements core.workflow().
//
// Reference: Workflow.java
func workflowFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		name             string
		origin           starlark.Value
		destination      starlark.Value
		authoring        starlark.Value
		transformations  *starlark.List
		originFiles      starlark.Value = starlark.None
		destinationFiles starlark.Value = starlark.None
		mode             string         = "SQUASH"
		reversibleCheck  starlark.Value = starlark.None
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"origin?", &origin,
		"destination?", &destination,
		"authoring?", &authoring,
		"transformations?", &transformations,
		"origin_files?", &originFiles,
		"destination_files?", &destinationFiles,
		"mode?", &mode,
		"reversible_check?", &reversibleCheck,
	); err != nil {
		return nil, err
	}

	// Parse workflow mode
	workflowMode, err := ParseWorkflowMode(mode)
	if err != nil {
		return nil, err
	}

	wf := &Workflow{
		name:        name,
		origin:      origin,
		destination: destination,
		authoring:   authoring,
		mode:        workflowMode,
	}

	// Handle origin_files
	wf.originFiles, err = wrapGlob(originFiles)
	if err != nil {
		return nil, fmt.Errorf("origin_files: %w", err)
	}

	// Handle destination_files
	wf.destinationFiles, err = wrapGlob(destinationFiles)
	if err != nil {
		return nil, fmt.Errorf("destination_files: %w", err)
	}

	// Handle reversible_check
	switch v := reversibleCheck.(type) {
	case starlark.NoneType:
		// Default: True for CHANGE_REQUEST mode, False otherwise
		wf.reversibleCheck = workflowMode == ModeChangeRequest
	case starlark.Bool:
		wf.reversibleCheck = bool(v)
	default:
		return nil, fmt.Errorf("reversible_check must be a bool, got %s", reversibleCheck.Type())
	}

	// Handle transformations
	if transformations != nil {
		wf.transformations = make([]Transformation, transformations.Len())
		for i := range transformations.Len() {
			item := transformations.Index(i)
			t, ok := item.(Transformation)
			if !ok {
				return nil, fmt.Errorf("transformations[%d] must be a transformation, got %s", i, item.Type())
			}
			wf.transformations[i] = t
		}
	}

	return wf, nil
}

// wrapGlob converts a starlark value to a Glob.
func wrapGlob(v starlark.Value) (*Glob, error) {
	switch val := v.(type) {
	case starlark.NoneType:
		return AllFiles(), nil
	case *Glob:
		return val, nil
	case *starlark.List:
		patterns := make([]string, val.Len())
		for i := range val.Len() {
			s, ok := starlark.AsString(val.Index(i))
			if !ok {
				return nil, fmt.Errorf("patterns must be strings, got %s", val.Index(i).Type())
			}
			patterns[i] = s
		}
		return NewGlob(patterns, nil)
	default:
		return nil, fmt.Errorf("expected glob or list of strings, got %s", v.Type())
	}
}

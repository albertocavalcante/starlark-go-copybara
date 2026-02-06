package core

import (
	"fmt"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Replace represents a search-and-replace transformation.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Replace.java
type Replace struct {
	before string
	after  string
	paths  []string
}

var _ Transformation = (*Replace)(nil)

// String implements starlark.Value.
func (r *Replace) String() string {
	return fmt.Sprintf("core.replace(%q, %q)", r.before, r.after)
}

// Type implements starlark.Value.
func (r *Replace) Type() string {
	return "replace"
}

// Freeze implements starlark.Value.
func (r *Replace) Freeze() {}

// Truth implements starlark.Value.
func (r *Replace) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (r *Replace) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: replace")
}

// Apply implements Transformation.
func (r *Replace) Apply(ctx *transform.Context) error {
	// TODO: Implement search and replace
	return nil
}

// Reverse implements Transformation.
func (r *Replace) Reverse() Transformation {
	return &Replace{
		before: r.after,
		after:  r.before,
		paths:  r.paths,
	}
}

// Describe implements Transformation.
func (r *Replace) Describe() string {
	return fmt.Sprintf("Replacing %q with %q", r.before, r.after)
}

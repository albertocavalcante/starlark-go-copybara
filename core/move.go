package core

import (
	"fmt"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Move represents a file move/rename transformation.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/CopyOrMove.java
type Move struct {
	before string
	after  string
}

var _ Transformation = (*Move)(nil)

// String implements starlark.Value.
func (m *Move) String() string {
	return fmt.Sprintf("core.move(%q, %q)", m.before, m.after)
}

// Type implements starlark.Value.
func (m *Move) Type() string {
	return "move"
}

// Freeze implements starlark.Value.
func (m *Move) Freeze() {}

// Truth implements starlark.Value.
func (m *Move) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (m *Move) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: move")
}

// Apply implements Transformation.
func (m *Move) Apply(ctx *transform.Context) error {
	// TODO: Implement file move
	return nil
}

// Reverse implements Transformation.
func (m *Move) Reverse() Transformation {
	return &Move{
		before: m.after,
		after:  m.before,
	}
}

// Describe implements Transformation.
func (m *Move) Describe() string {
	return fmt.Sprintf("Moving %s to %s", m.before, m.after)
}

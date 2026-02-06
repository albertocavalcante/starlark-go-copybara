package core

import (
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Transformation is the interface for all code transformations.
type Transformation interface {
	// Apply applies the transformation to the given context.
	Apply(ctx *transform.Context) error

	// Reverse returns the reverse of this transformation.
	Reverse() Transformation

	// Describe returns a human-readable description.
	Describe() string
}

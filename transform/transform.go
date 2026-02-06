// Package transform provides the transformation execution context and shared interfaces.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Transformation.java
package transform

// Transformation is the interface for all transformations.
// This is the shared interface used by both core and metadata transformations.
type Transformation interface {
	// Apply applies the transformation to the given context.
	Apply(ctx *Context) error

	// Reverse returns the reverse of this transformation.
	// Note: Some transformations may return a noop or error transformation
	// if they are not reversible.
	Reverse() Transformation

	// Describe returns a human-readable description of the transformation.
	Describe() string
}

// NoopTransformation is a transformation that does nothing.
// It can be used as the reverse of non-reversible transformations.
type NoopTransformation struct {
	original Transformation
}

// NewNoopTransformation creates a new noop transformation.
func NewNoopTransformation(original Transformation) *NoopTransformation {
	return &NoopTransformation{original: original}
}

// Apply implements Transformation.
func (n *NoopTransformation) Apply(ctx *Context) error {
	return nil
}

// Reverse implements Transformation.
func (n *NoopTransformation) Reverse() Transformation {
	if n.original != nil {
		return n.original
	}
	return n
}

// Describe implements Transformation.
func (n *NoopTransformation) Describe() string {
	return "noop"
}

// ErrorTransformation is a transformation that returns an error when applied.
// It can be used for non-reversible transformations when reversed.
type ErrorTransformation struct {
	err      error
	original Transformation
}

// NewErrorTransformation creates a new error transformation.
func NewErrorTransformation(err error, original Transformation) *ErrorTransformation {
	return &ErrorTransformation{err: err, original: original}
}

// Apply implements Transformation.
func (e *ErrorTransformation) Apply(ctx *Context) error {
	return e.err
}

// Reverse implements Transformation.
func (e *ErrorTransformation) Reverse() Transformation {
	if e.original != nil {
		return e.original
	}
	return e
}

// Describe implements Transformation.
func (e *ErrorTransformation) Describe() string {
	return "error transformation"
}

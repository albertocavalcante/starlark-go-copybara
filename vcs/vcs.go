// Package vcs provides version control system abstractions.
package vcs

import (
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Repository represents a version control repository.
type Repository interface {
	// URL returns the repository URL.
	URL() string

	// Ref returns the current reference.
	Ref() string

	// Checkout checks out the given reference.
	Checkout(ref string) error
}

// Change is an alias to transform.Change for VCS operations.
// This maintains backward compatibility while consolidating to a single Change type.
type Change = transform.Change

// Origin is a source repository for migrations.
type Origin interface {
	Repository

	// Changes returns changes since the given baseline reference.
	Changes(baseline string) ([]*Change, error)
}

// Destination is a target repository for migrations.
type Destination interface {
	Repository

	// Write writes changes to the destination.
	Write(changes []*Change) error
}

// Package vcs provides version control system abstractions.
package vcs

// Repository represents a version control repository.
type Repository interface {
	// URL returns the repository URL.
	URL() string

	// Ref returns the current reference.
	Ref() string

	// Checkout checks out the given reference.
	Checkout(ref string) error
}

// Change represents a version control change (commit).
type Change struct {
	// Ref is the change reference (commit hash).
	Ref string

	// Author is the change author.
	Author string

	// Message is the commit message.
	Message string

	// Files is the list of changed files.
	Files []string
}

// Origin is a source repository for migrations.
type Origin interface {
	Repository

	// Changes returns changes since the given baseline reference.
	Changes(baseline string) ([]Change, error)
}

// Destination is a target repository for migrations.
type Destination interface {
	Repository

	// Write writes changes to the destination.
	Write(changes []Change) error
}

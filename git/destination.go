package git

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Destination represents a Git repository destination.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitDestination.java
type Destination struct {
	url   string
	push  string
	fetch string
}

// String implements starlark.Value.
func (d *Destination) String() string {
	return fmt.Sprintf("git.destination(url = %q)", d.url)
}

// Type implements starlark.Value.
func (d *Destination) Type() string {
	return "git.destination"
}

// Freeze implements starlark.Value.
func (d *Destination) Freeze() {}

// Truth implements starlark.Value.
func (d *Destination) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (d *Destination) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.destination")
}

// URL returns the repository URL.
func (d *Destination) URL() string {
	return d.url
}

// Push returns the push branch.
func (d *Destination) Push() string {
	return d.push
}

// Fetch returns the fetch branch.
func (d *Destination) Fetch() string {
	return d.fetch
}

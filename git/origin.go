package git

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Origin represents a Git repository origin.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitOrigin.java
type Origin struct {
	url string
	ref string
}

// String implements starlark.Value.
func (o *Origin) String() string {
	if o.ref != "" {
		return fmt.Sprintf("git.origin(url = %q, ref = %q)", o.url, o.ref)
	}
	return fmt.Sprintf("git.origin(url = %q)", o.url)
}

// Type implements starlark.Value.
func (o *Origin) Type() string {
	return "git.origin"
}

// Freeze implements starlark.Value.
func (o *Origin) Freeze() {}

// Truth implements starlark.Value.
func (o *Origin) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (o *Origin) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.origin")
}

// URL returns the repository URL.
func (o *Origin) URL() string {
	return o.url
}

// Ref returns the Git reference.
func (o *Origin) Ref() string {
	return o.ref
}

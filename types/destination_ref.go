package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Ensure DestinationRef implements required interfaces.
var (
	_ starlark.Value    = (*DestinationRef)(nil)
	_ starlark.HasAttrs = (*DestinationRef)(nil)
)

// DestinationRef represents a reference to a specific point in a destination repository.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Destination.java
type DestinationRef struct {
	ref string
	url string
}

// NewDestinationRef creates a new DestinationRef.
func NewDestinationRef(ref, url string) *DestinationRef {
	return &DestinationRef{
		ref: ref,
		url: url,
	}
}

// String implements starlark.Value.
func (d *DestinationRef) String() string {
	return fmt.Sprintf("destination_ref<%s>", d.ref)
}

// Type implements starlark.Value.
func (d *DestinationRef) Type() string {
	return "destination_ref"
}

// Freeze implements starlark.Value.
func (d *DestinationRef) Freeze() {}

// Truth implements starlark.Value.
func (d *DestinationRef) Truth() starlark.Bool {
	return starlark.Bool(d.ref != "")
}

// Hash implements starlark.Value.
func (d *DestinationRef) Hash() (uint32, error) {
	return starlark.String(d.ref).Hash()
}

// Ref returns the reference string (commit hash, tag, branch, etc.).
func (d *DestinationRef) Ref() string {
	return d.ref
}

// URL returns the URL of the destination repository.
func (d *DestinationRef) URL() string {
	return d.url
}

// Attr implements starlark.HasAttrs.
func (d *DestinationRef) Attr(name string) (starlark.Value, error) {
	switch name {
	case "ref":
		return starlark.String(d.ref), nil
	case "url":
		return starlark.String(d.url), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (d *DestinationRef) AttrNames() []string {
	return []string{"ref", "url"}
}

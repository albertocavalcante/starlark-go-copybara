package types

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Attribute names for origin_ref.
const (
	attrRef = "ref"
	attrURL = "url"
)

// Ensure OriginRef implements required interfaces.
var (
	_ starlark.Value    = (*OriginRef)(nil)
	_ starlark.HasAttrs = (*OriginRef)(nil)
)

// OriginRef represents a reference to a specific point in an origin repository.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Origin.java
type OriginRef struct {
	ref string
	url string
}

// NewOriginRef creates a new OriginRef.
func NewOriginRef(ref, url string) *OriginRef {
	return &OriginRef{
		ref: ref,
		url: url,
	}
}

// String implements starlark.Value.
func (o *OriginRef) String() string {
	return fmt.Sprintf("origin_ref<%s>", o.ref)
}

// Type implements starlark.Value.
func (o *OriginRef) Type() string {
	return "origin_ref"
}

// Freeze implements starlark.Value.
func (o *OriginRef) Freeze() {}

// Truth implements starlark.Value.
func (o *OriginRef) Truth() starlark.Bool {
	return starlark.Bool(o.ref != "")
}

// Hash implements starlark.Value.
func (o *OriginRef) Hash() (uint32, error) {
	return starlark.String(o.ref).Hash()
}

// Ref returns the reference string (commit hash, tag, branch, etc.).
func (o *OriginRef) Ref() string {
	return o.ref
}

// URL returns the URL of the origin repository.
func (o *OriginRef) URL() string {
	return o.url
}

// Attr implements starlark.HasAttrs.
func (o *OriginRef) Attr(name string) (starlark.Value, error) {
	switch name {
	case attrRef:
		return starlark.String(o.ref), nil
	case attrURL:
		return starlark.String(o.url), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (o *OriginRef) AttrNames() []string {
	return []string{attrRef, attrURL}
}

// Package types provides core types for Copybara configurations.
package types

import (
	"fmt"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
)

// Ensure Path implements required interfaces.
var (
	_ starlark.Value    = (*Path)(nil)
	_ starlark.HasAttrs = (*Path)(nil)
)

// Path represents a file path in the working tree.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/CheckoutPath.java
type Path struct {
	path string
}

// NewPath creates a new Path from a string.
func NewPath(p string) *Path {
	return &Path{path: filepath.Clean(p)}
}

// String implements starlark.Value.
func (p *Path) String() string {
	return p.path
}

// Type implements starlark.Value.
func (p *Path) Type() string {
	return "path"
}

// Freeze implements starlark.Value.
func (p *Path) Freeze() {}

// Truth implements starlark.Value.
func (p *Path) Truth() starlark.Bool {
	return starlark.Bool(p.path != "")
}

// Hash implements starlark.Value.
func (p *Path) Hash() (uint32, error) {
	return starlark.String(p.path).Hash()
}

// Join returns a new path with the given segments appended.
func (p *Path) Join(segments ...string) *Path {
	parts := append([]string{p.path}, segments...)
	return NewPath(filepath.Join(parts...))
}

// Parent returns the parent directory.
func (p *Path) Parent() *Path {
	return NewPath(filepath.Dir(p.path))
}

// Base returns the last element of the path.
func (p *Path) Base() string {
	return filepath.Base(p.path)
}

// ResolveSibling returns a new path with the same parent but a different name.
func (p *Path) ResolveSibling(name string) *Path {
	parent := filepath.Dir(p.path)
	return NewPath(filepath.Join(parent, name))
}

// StartsWith returns true if this path starts with the given prefix.
func (p *Path) StartsWith(prefix string) bool {
	cleanPrefix := filepath.Clean(prefix)
	cleanPath := p.path

	// Handle exact match
	if cleanPath == cleanPrefix {
		return true
	}

	// Ensure we match at path boundaries
	if !strings.HasSuffix(cleanPrefix, string(filepath.Separator)) {
		cleanPrefix += string(filepath.Separator)
	}
	return strings.HasPrefix(cleanPath+string(filepath.Separator), cleanPrefix)
}

// Relativize returns a relative path from this path to the given path.
// Returns an error if the other path is not under this path.
func (p *Path) Relativize(other *Path) (*Path, error) {
	rel, err := filepath.Rel(p.path, other.path)
	if err != nil {
		return nil, err
	}
	return NewPath(rel), nil
}

// Attr implements starlark.HasAttrs.
func (p *Path) Attr(name string) (starlark.Value, error) {
	switch name {
	case "path":
		return starlark.String(p.path), nil
	case "name":
		return starlark.String(p.Base()), nil
	case "parent":
		return p.Parent(), nil
	case "resolve_sibling":
		return starlark.NewBuiltin("path.resolve_sibling", p.resolveSiblingBuiltin), nil
	case "starts_with":
		return starlark.NewBuiltin("path.starts_with", p.startsWithBuiltin), nil
	case "relativize":
		return starlark.NewBuiltin("path.relativize", p.relativizeBuiltin), nil
	case "join":
		return starlark.NewBuiltin("path.join", p.joinBuiltin), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (p *Path) AttrNames() []string {
	return []string{"join", "name", "parent", "path", "relativize", "resolve_sibling", "starts_with"}
}

// resolveSiblingBuiltin is the Starlark builtin for resolve_sibling.
func (p *Path) resolveSiblingBuiltin(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	return p.ResolveSibling(name), nil
}

// startsWithBuiltin is the Starlark builtin for starts_with.
func (p *Path) startsWithBuiltin(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var prefix string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "prefix", &prefix); err != nil {
		return nil, err
	}
	return starlark.Bool(p.StartsWith(prefix)), nil
}

// relativizeBuiltin is the Starlark builtin for relativize.
func (p *Path) relativizeBuiltin(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var other *Path
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "other", &other); err != nil {
		return nil, err
	}
	result, err := p.Relativize(other)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// joinBuiltin is the Starlark builtin for join.
func (p *Path) joinBuiltin(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	segments := make([]string, len(args))
	for i, arg := range args {
		s, ok := starlark.AsString(arg)
		if !ok {
			return nil, fmt.Errorf("join: expected string, got %s", arg.Type())
		}
		segments[i] = s
	}
	return p.Join(segments...), nil
}

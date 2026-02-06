// Package types provides core types for Copybara configurations.
package types

import (
	"fmt"
	"path/filepath"

	"go.starlark.net/starlark"
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

// Attr implements starlark.HasAttrs.
func (p *Path) Attr(name string) (starlark.Value, error) {
	switch name {
	case "path":
		return starlark.String(p.path), nil
	case "name":
		return starlark.String(p.Base()), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (p *Path) AttrNames() []string {
	return []string{"path", "name"}
}

// Glob represents a glob pattern for matching files.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/util/Glob.java
type Glob struct {
	include []string
	exclude []string
}

// NewGlob creates a new Glob from include and exclude patterns.
func NewGlob(include, exclude []string) *Glob {
	return &Glob{
		include: include,
		exclude: exclude,
	}
}

// String implements starlark.Value.
func (g *Glob) String() string {
	return fmt.Sprintf("glob(%v)", g.include)
}

// Type implements starlark.Value.
func (g *Glob) Type() string {
	return "glob"
}

// Freeze implements starlark.Value.
func (g *Glob) Freeze() {}

// Truth implements starlark.Value.
func (g *Glob) Truth() starlark.Bool {
	return starlark.Bool(len(g.include) > 0)
}

// Hash implements starlark.Value.
func (g *Glob) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: glob")
}

// Package core provides the core Copybara module with transformations.
package core

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

// Glob represents a set of relative filepaths in the Copybara workdir.
//
// A glob matches files based on include patterns, optionally excluding
// files that match exclude patterns.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/util/Glob.java
type Glob struct {
	include []string
	exclude *Glob
	frozen  bool
}

// Ensure Glob implements required interfaces.
var (
	_ starlark.Value     = (*Glob)(nil)
	_ starlark.HasBinary = (*Glob)(nil)
)

// AllFiles returns a Glob that matches all files.
func AllFiles() *Glob {
	return &Glob{
		include: []string{"**"},
	}
}

// NewGlob creates a new Glob from include and exclude patterns.
func NewGlob(include, exclude []string) (*Glob, error) {
	if len(include) == 0 {
		return nil, fmt.Errorf("glob include patterns cannot be empty")
	}

	for _, pattern := range include {
		if err := validateGlobPattern(pattern); err != nil {
			return nil, fmt.Errorf("invalid include pattern %q: %w", pattern, err)
		}
	}

	g := &Glob{
		include: slices.Clone(include),
	}

	if len(exclude) > 0 {
		for _, pattern := range exclude {
			if err := validateGlobPattern(pattern); err != nil {
				return nil, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
			}
		}
		g.exclude = &Glob{
			include: slices.Clone(exclude),
		}
	}

	return g, nil
}

// validateGlobPattern validates that a glob pattern is well-formed.
func validateGlobPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	if strings.HasPrefix(pattern, "/") {
		return fmt.Errorf("pattern cannot start with /")
	}
	// Check for valid glob syntax by trying to match against empty string
	_, err := filepath.Match(strings.ReplaceAll(pattern, "**", "*"), "")
	if err != nil {
		return err
	}
	return nil
}

// String implements starlark.Value.
func (g *Glob) String() string {
	var sb strings.Builder
	sb.WriteString("glob(include = [")
	for i, p := range g.include {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%q", p))
	}
	sb.WriteString("]")

	if g.exclude != nil && len(g.exclude.include) > 0 {
		sb.WriteString(", exclude = [")
		for i, p := range g.exclude.include {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%q", p))
		}
		sb.WriteString("]")
	}

	sb.WriteString(")")
	return sb.String()
}

// Type implements starlark.Value.
func (g *Glob) Type() string {
	return "glob"
}

// Freeze implements starlark.Value.
func (g *Glob) Freeze() {
	g.frozen = true
	if g.exclude != nil {
		g.exclude.Freeze()
	}
}

// Truth implements starlark.Value.
func (g *Glob) Truth() starlark.Bool {
	return starlark.Bool(len(g.include) > 0)
}

// Hash implements starlark.Value.
func (g *Glob) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: glob")
}

// Binary implements starlark.HasBinary for + and - operators.
func (g *Glob) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
	if side == starlark.Right {
		return nil, nil // let the other operand handle it
	}

	switch op {
	case syntax.PLUS:
		// Union of two globs
		switch other := y.(type) {
		case *Glob:
			return Union(g, other), nil
		case *starlark.List:
			otherGlob, err := globFromList(other)
			if err != nil {
				return nil, err
			}
			return Union(g, otherGlob), nil
		default:
			return nil, fmt.Errorf("cannot concatenate glob with %s", y.Type())
		}

	case syntax.MINUS:
		// Difference of two globs
		switch other := y.(type) {
		case *Glob:
			return Difference(g, other), nil
		case *starlark.List:
			otherGlob, err := globFromList(other)
			if err != nil {
				return nil, err
			}
			return Difference(g, otherGlob), nil
		default:
			return nil, fmt.Errorf("cannot subtract %s from glob", y.Type())
		}

	default:
		return nil, fmt.Errorf("glob does not support %s", op)
	}
}

// globFromList converts a starlark list of strings to a Glob.
func globFromList(list *starlark.List) (*Glob, error) {
	patterns := make([]string, list.Len())
	for i := range list.Len() {
		s, ok := starlark.AsString(list.Index(i))
		if !ok {
			return nil, fmt.Errorf("glob list must contain strings, got %s", list.Index(i).Type())
		}
		patterns[i] = s
	}
	return NewGlob(patterns, nil)
}

// Union computes the set union of two Globs.
func Union(g1, g2 *Glob) *Glob {
	// If both have the same exclude, we can merge includes
	if globsEqual(g1.exclude, g2.exclude) {
		combined := make([]string, 0, len(g1.include)+len(g2.include))
		combined = append(combined, g1.include...)
		combined = append(combined, g2.include...)
		return &Glob{
			include: combined,
			exclude: g1.exclude,
		}
	}

	// Otherwise, create a new glob that includes both
	combined := make([]string, 0, len(g1.include)+len(g2.include))
	combined = append(combined, g1.include...)
	combined = append(combined, g2.include...)
	return &Glob{
		include: combined,
	}
}

// Difference computes the set difference of two Globs.
func Difference(g1, g2 *Glob) *Glob {
	if g1.exclude == nil {
		return &Glob{
			include: slices.Clone(g1.include),
			exclude: g2,
		}
	}
	return &Glob{
		include: slices.Clone(g1.include),
		exclude: Union(g1.exclude, g2),
	}
}

// globsEqual checks if two globs are equal.
func globsEqual(g1, g2 *Glob) bool {
	if g1 == nil && g2 == nil {
		return true
	}
	if g1 == nil || g2 == nil {
		return false
	}
	if !slices.Equal(g1.include, g2.include) {
		return false
	}
	return globsEqual(g1.exclude, g2.exclude)
}

// Include returns the include patterns.
func (g *Glob) Include() []string {
	return g.include
}

// Exclude returns the exclude Glob (may be nil).
func (g *Glob) Exclude() *Glob {
	return g.exclude
}

// ExcludePatterns returns the exclude patterns as a flat list.
func (g *Glob) ExcludePatterns() []string {
	if g.exclude == nil {
		return nil
	}
	return g.exclude.include
}

// Matches checks if a path matches this glob.
func (g *Glob) Matches(path string) bool {
	// Check if path matches any include pattern
	matched := false
	for _, pattern := range g.include {
		if matchGlobPattern(pattern, path) {
			matched = true
			break
		}
	}

	if !matched {
		return false
	}

	// Check if path matches any exclude pattern
	if g.exclude != nil {
		for _, pattern := range g.exclude.include {
			if matchGlobPattern(pattern, path) {
				return false
			}
		}
	}

	return true
}

// matchGlobPattern matches a path against a glob pattern.
// Supports ** for recursive matching.
func matchGlobPattern(pattern, path string) bool {
	// Handle ** (double star) for recursive matching
	if strings.Contains(pattern, "**") {
		return matchDoubleStarPattern(pattern, path)
	}

	// Use standard filepath.Match for simple patterns
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// matchDoubleStarPattern handles ** glob patterns.
func matchDoubleStarPattern(pattern, path string) bool {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 1 {
		// No ** in pattern, use simple match
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// Handle pattern starting with **
	if parts[0] == "" {
		// Pattern is "**" or starts with "**"
		if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") {
			return true // ** matches everything
		}

		suffix := parts[1]
		if strings.HasPrefix(suffix, "/") {
			suffix = suffix[1:]
		}

		// Try matching suffix against path and all subdirectories
		pathParts := strings.Split(path, "/")
		for i := range len(pathParts) {
			subPath := strings.Join(pathParts[i:], "/")
			if matchGlobPattern(suffix, subPath) {
				return true
			}
		}
		return false
	}

	// Handle pattern ending with **
	if parts[len(parts)-1] == "" || parts[len(parts)-1] == "/" {
		prefix := strings.TrimSuffix(parts[0], "/")
		return strings.HasPrefix(path, prefix+"/") || path == prefix
	}

	// Handle pattern with ** in the middle
	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	if !strings.HasPrefix(path, prefix) {
		if prefix != "" {
			return false
		}
	}

	remaining := strings.TrimPrefix(path, prefix)
	remaining = strings.TrimPrefix(remaining, "/")

	// Try matching suffix against remaining path and all subdirectories
	remainingParts := strings.Split(remaining, "/")
	for i := range len(remainingParts) {
		subPath := strings.Join(remainingParts[i:], "/")
		if matchGlobPattern(suffix, subPath) {
			return true
		}
	}

	return false
}

// Roots returns a set of root paths that contain all files that could match this glob.
func (g *Glob) Roots() []string {
	roots := make([]string, 0, len(g.include))

	for _, pattern := range g.include {
		root := computeRoot(pattern)
		roots = append(roots, root)
	}

	// Remove duplicates and sort
	slices.Sort(roots)
	roots = slices.Compact(roots)

	// Remove roots that are subpaths of other roots
	if slices.Contains(roots, "") {
		return []string{""}
	}

	filtered := make([]string, 0, len(roots))
	for i, root := range roots {
		isSubpath := false
		for j, other := range roots {
			if i != j && strings.HasPrefix(root, other+"/") {
				isSubpath = true
				break
			}
		}
		if !isSubpath {
			filtered = append(filtered, root)
		}
	}

	return filtered
}

// computeRoot extracts the root directory from a glob pattern.
func computeRoot(pattern string) string {
	parts := strings.Split(pattern, "/")
	var root []string

	for _, part := range parts {
		if strings.ContainsAny(part, "*?[{") {
			break
		}
		root = append(root, part)
	}

	return strings.Join(root, "/")
}

// IsAllFiles returns true if this glob matches all files.
func (g *Glob) IsAllFiles() bool {
	return len(g.include) == 1 && g.include[0] == "**" && g.exclude == nil
}

// globFn implements the glob() Starlark function.
func globFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var include *starlark.List
	var exclude *starlark.List

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"include", &include,
		"exclude?", &exclude,
	); err != nil {
		return nil, err
	}

	includePatterns := make([]string, include.Len())
	for i := range include.Len() {
		s, ok := starlark.AsString(include.Index(i))
		if !ok {
			return nil, fmt.Errorf("include patterns must be strings, got %s", include.Index(i).Type())
		}
		includePatterns[i] = s
	}

	var excludePatterns []string
	if exclude != nil {
		excludePatterns = make([]string, exclude.Len())
		for i := range exclude.Len() {
			s, ok := starlark.AsString(exclude.Index(i))
			if !ok {
				return nil, fmt.Errorf("exclude patterns must be strings, got %s", exclude.Index(i).Type())
			}
			excludePatterns[i] = s
		}
	}

	return NewGlob(includePatterns, excludePatterns)
}

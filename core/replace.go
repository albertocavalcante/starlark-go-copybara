package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Replace represents a search-and-replace transformation.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/Replace.java
type Replace struct {
	before string
	after  string
	paths  *Glob
}

var _ Transformation = (*Replace)(nil)

// String implements starlark.Value.
func (r *Replace) String() string {
	if r.paths != nil && !r.paths.IsAllFiles() {
		return fmt.Sprintf("core.replace(%q, %q, paths = %s)", r.before, r.after, r.paths)
	}
	return fmt.Sprintf("core.replace(%q, %q)", r.before, r.after)
}

// Type implements starlark.Value.
func (r *Replace) Type() string {
	return "replace"
}

// Freeze implements starlark.Value.
func (r *Replace) Freeze() {}

// Truth implements starlark.Value.
func (r *Replace) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (r *Replace) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: replace")
}

// Apply implements Transformation.
func (r *Replace) Apply(ctx *transform.Context) error {
	if ctx.WorkDir == "" {
		return fmt.Errorf("workdir is required for replace transformation")
	}

	return filepath.WalkDir(ctx.WorkDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and symlinks
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, err := filepath.Rel(ctx.WorkDir, path)
		if err != nil {
			return err
		}

		// Check if file matches glob
		if r.paths != nil && !r.paths.Matches(relPath) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", relPath, err)
		}

		// Perform replacement
		newContent := strings.ReplaceAll(string(content), r.before, r.after)

		// Only write if content changed
		if newContent != string(content) {
			if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
				return fmt.Errorf("failed to write file %q: %w", relPath, err)
			}
		}

		return nil
	})
}

// Reverse implements Transformation.
func (r *Replace) Reverse() transform.Transformation {
	return &Replace{
		before: r.after,
		after:  r.before,
		paths:  r.paths,
	}
}

// Describe implements Transformation.
func (r *Replace) Describe() string {
	return fmt.Sprintf("Replacing %q with %q", r.before, r.after)
}

// Before returns the search string.
func (r *Replace) Before() string {
	return r.before
}

// After returns the replacement string.
func (r *Replace) After() string {
	return r.after
}

// Paths returns the glob filter.
func (r *Replace) Paths() *Glob {
	return r.paths
}

package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Remove represents a file removal transformation.
//
// This transformation is meant to be used inside core.transform for
// reversing core.copy like transforms.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/Remove.java
type Remove struct {
	paths *Glob
}

var _ Transformation = (*Remove)(nil)

// String implements starlark.Value.
func (r *Remove) String() string {
	return fmt.Sprintf("core.remove(%s)", r.paths)
}

// Type implements starlark.Value.
func (r *Remove) Type() string {
	return "remove"
}

// Freeze implements starlark.Value.
func (r *Remove) Freeze() {}

// Truth implements starlark.Value.
func (r *Remove) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (r *Remove) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: remove")
}

// Apply implements Transformation.
func (r *Remove) Apply(ctx *transform.Context) error {
	if ctx.WorkDir == "" {
		return fmt.Errorf("workdir is required for remove transformation")
	}

	var filesToRemove []string

	// Walk the workdir and collect files matching the glob
	err := filepath.WalkDir(ctx.WorkDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(ctx.WorkDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory
		if relPath == "." {
			return nil
		}

		if r.paths.Matches(relPath) {
			filesToRemove = append(filesToRemove, path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// Remove files in reverse order (deepest first) to handle directories
	for i := len(filesToRemove) - 1; i >= 0; i-- {
		path := filesToRemove[i]
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %q: %w", path, err)
		}
	}

	return nil
}

// Reverse implements Transformation.
// Remove is not reversible.
func (r *Remove) Reverse() transform.Transformation {
	return transform.NewNoopTransformation(r)
}

// Describe implements Transformation.
func (r *Remove) Describe() string {
	return fmt.Sprintf("Removing files matching %s", r.paths)
}

// Paths returns the glob of paths to remove.
func (r *Remove) Paths() *Glob {
	return r.paths
}

// removeFn implements core.remove().
func removeFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var paths starlark.Value

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"paths", &paths,
	); err != nil {
		return nil, err
	}

	remove := &Remove{}

	// Handle paths parameter
	switch v := paths.(type) {
	case *Glob:
		remove.paths = v
	case *starlark.List:
		patterns := make([]string, v.Len())
		for i := range v.Len() {
			s, ok := starlark.AsString(v.Index(i))
			if !ok {
				return nil, fmt.Errorf("paths must be strings, got %s", v.Index(i).Type())
			}
			patterns[i] = s
		}
		var err error
		remove.paths, err = NewGlob(patterns, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("paths must be a glob or list of strings, got %s", paths.Type())
	}

	return remove, nil
}

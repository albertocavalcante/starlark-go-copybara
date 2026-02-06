package core

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Copy represents a file copy transformation.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/CopyOrMove.java
type Copy struct {
	before    string
	after     string
	paths     *Glob
	overwrite bool
}

var _ Transformation = (*Copy)(nil)

// String implements starlark.Value.
func (c *Copy) String() string {
	if c.paths != nil && !c.paths.IsAllFiles() {
		return fmt.Sprintf("core.copy(%q, %q, paths = %s)", c.before, c.after, c.paths)
	}
	return fmt.Sprintf("core.copy(%q, %q)", c.before, c.after)
}

// Type implements starlark.Value.
func (c *Copy) Type() string {
	return "copy"
}

// Freeze implements starlark.Value.
func (c *Copy) Freeze() {}

// Truth implements starlark.Value.
func (c *Copy) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (c *Copy) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: copy")
}

// Apply implements Transformation.
func (c *Copy) Apply(ctx *transform.Context) error {
	if ctx.WorkDir == "" {
		return fmt.Errorf("workdir is required for copy transformation")
	}

	beforePath := filepath.Join(ctx.WorkDir, c.before)
	afterPath := filepath.Join(ctx.WorkDir, c.after)

	// Check if source exists
	info, err := os.Stat(beforePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("source path %q does not exist", c.before)
	}
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Handle directory copy
	if info.IsDir() {
		return c.copyDir(beforePath, afterPath)
	}

	// Handle file copy
	return c.copyFile(beforePath, afterPath)
}

// copyDir recursively copies a directory.
func (c *Copy) copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Check if path matches glob patterns
		if c.paths != nil && !c.paths.IsAllFiles() {
			if !d.IsDir() && !c.paths.Matches(relPath) {
				return nil
			}
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o750)
		}

		return c.copyFile(path, dstPath)
	})
}

// copyFile copies a single file.
func (c *Copy) copyFile(src, dst string) error {
	// Check if destination exists
	if !c.overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination %q already exists (use overwrite=True to overwrite)", dst)
		}
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o750); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Open source file
	srcFile, err := os.Open(src) //nolint:gosec // src is validated by caller
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode()) //nolint:gosec // dst is validated by caller
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	return nil
}

// Reverse implements Transformation.
// Copy is not automatically reversible since it doesn't remove the original.
// The reverse is a Remove transformation.
func (c *Copy) Reverse() transform.Transformation {
	// Return a remove transformation that removes the copied files
	return &Remove{
		paths: &Glob{
			include: []string{c.after, c.after + "/**"},
		},
	}
}

// Describe implements Transformation.
func (c *Copy) Describe() string {
	return fmt.Sprintf("Copying %s to %s", c.before, c.after)
}

// Before returns the source path.
func (c *Copy) Before() string {
	return c.before
}

// After returns the destination path.
func (c *Copy) After() string {
	return c.after
}

// Paths returns the glob filter (may be nil).
func (c *Copy) Paths() *Glob {
	return c.paths
}

// copyFn implements core.copy().
func copyFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		before    string
		after     string
		paths     starlark.Value = starlark.None
		overwrite bool
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"before", &before,
		"after", &after,
		"paths?", &paths,
		"overwrite?", &overwrite,
	); err != nil {
		return nil, err
	}

	if before == after {
		return nil, fmt.Errorf("copying from the same folder to the same folder is a noop")
	}

	cp := &Copy{
		before:    before,
		after:     after,
		overwrite: overwrite,
	}

	// Handle paths parameter
	switch v := paths.(type) {
	case starlark.NoneType:
		cp.paths = AllFiles()
	case *Glob:
		cp.paths = v
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
		cp.paths, err = NewGlob(patterns, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("paths must be a glob or list of strings, got %s", paths.Type())
	}

	return cp, nil
}

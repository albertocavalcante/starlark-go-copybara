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

// Move represents a file move/rename transformation.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/CopyOrMove.java
type Move struct {
	before    string
	after     string
	paths     *Glob
	overwrite bool
}

var _ Transformation = (*Move)(nil)

// String implements starlark.Value.
func (m *Move) String() string {
	if m.paths != nil && !m.paths.IsAllFiles() {
		return fmt.Sprintf("core.move(%q, %q, paths = %s)", m.before, m.after, m.paths)
	}
	return fmt.Sprintf("core.move(%q, %q)", m.before, m.after)
}

// Type implements starlark.Value.
func (m *Move) Type() string {
	return "move"
}

// Freeze implements starlark.Value.
func (m *Move) Freeze() {}

// Truth implements starlark.Value.
func (m *Move) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (m *Move) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: move")
}

// Apply implements Transformation.
func (m *Move) Apply(ctx *transform.Context) error {
	if ctx.WorkDir == "" {
		return fmt.Errorf("workdir is required for move transformation")
	}

	beforePath := filepath.Join(ctx.WorkDir, m.before)
	afterPath := filepath.Join(ctx.WorkDir, m.after)

	// Check if source exists
	info, err := os.Stat(beforePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("source path %q does not exist", m.before)
	}
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// Handle directory move
	if info.IsDir() {
		return m.moveDir(beforePath, afterPath)
	}

	// Handle file move
	return m.moveFile(beforePath, afterPath)
}

// moveDir recursively moves a directory.
func (m *Move) moveDir(src, dst string) error {
	// First, copy all files that match the glob
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Check if path matches glob patterns
		if m.paths != nil && !m.paths.IsAllFiles() {
			if !d.IsDir() && !m.paths.Matches(relPath) {
				return nil
			}
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		// Move file
		if err := m.moveFile(path, dstPath); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Remove empty directories
	return m.removeEmptyDirs(src)
}

// moveFile moves a single file.
func (m *Move) moveFile(src, dst string) error {
	// Check if destination exists
	if !m.overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("destination %q already exists (use overwrite=True to overwrite)", dst)
		}
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Try rename first (atomic move on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// Fall back to copy+delete for cross-filesystem moves
	if err := m.copyAndDelete(src, dst); err != nil {
		return err
	}

	return nil
}

// copyAndDelete copies a file and deletes the source.
func (m *Move) copyAndDelete(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Close files before deleting
	srcFile.Close()
	dstFile.Close()

	// Delete source
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("failed to remove source file: %w", err)
	}

	return nil
}

// removeEmptyDirs removes empty directories recursively.
func (m *Move) removeEmptyDirs(dir string) error {
	// Walk directories in reverse order (deepest first)
	var dirs []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			dirs = append(dirs, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Try to remove directories in reverse order
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := os.Remove(dirs[i]); err != nil {
			// Ignore errors - directory may not be empty
			continue
		}
	}

	return nil
}

// Reverse implements Transformation.
func (m *Move) Reverse() transform.Transformation {
	if m.overwrite {
		// Not reversible when overwrite is set
		return transform.NewNoopTransformation(m)
	}
	return &Move{
		before:    m.after,
		after:     m.before,
		paths:     m.paths,
		overwrite: false,
	}
}

// Describe implements Transformation.
func (m *Move) Describe() string {
	return fmt.Sprintf("Moving %s to %s", m.before, m.after)
}

// Before returns the source path.
func (m *Move) Before() string {
	return m.before
}

// After returns the destination path.
func (m *Move) After() string {
	return m.after
}

// Paths returns the glob filter (may be nil).
func (m *Move) Paths() *Glob {
	return m.paths
}

// Overwrite returns whether overwrite is enabled.
func (m *Move) Overwrite() bool {
	return m.overwrite
}

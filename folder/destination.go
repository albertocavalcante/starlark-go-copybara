package folder

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/albertocavalcante/starlark-go-copybara/vcs"
)

// Compile-time interface verification.
var _ vcs.Destination = (*DestinationImpl)(nil)

// DestinationImpl implements the vcs.Destination interface for local folders.
type DestinationImpl struct {
	destination *Destination
	fs          FileSystem
}

// NewDestinationImpl creates a new DestinationImpl from a Destination configuration.
func NewDestinationImpl(dest *Destination) *DestinationImpl {
	return &DestinationImpl{
		destination: dest,
		fs:          NewOSFileSystem(),
	}
}

// WithFileSystem sets a custom filesystem (useful for testing).
func (d *DestinationImpl) WithFileSystem(fs FileSystem) *DestinationImpl {
	d.fs = fs
	return d
}

// URL returns the folder path as the URL.
func (d *DestinationImpl) URL() string {
	return d.path()
}

// path returns the resolved path, defaulting to current working directory.
func (d *DestinationImpl) path() string {
	if d.destination.path != "" {
		return d.destination.path
	}
	// Default to current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// Ref returns an empty string (folders don't have refs).
func (d *DestinationImpl) Ref() string {
	return ""
}

// Checkout is a no-op for folder destinations.
func (d *DestinationImpl) Checkout(ref string) error {
	return nil
}

// Write writes changes to the destination folder.
func (d *DestinationImpl) Write(changes []*vcs.Change) error {
	path := d.path()

	// Ensure destination directory exists
	if err := d.fs.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", path, err)
	}

	// Write is typically called after transformations have already placed
	// files in a working directory. For folder destinations, this is usually
	// a no-op since files are written directly during transformation.
	return nil
}

// WriteFile writes a single file to the destination.
func (d *DestinationImpl) WriteFile(relativePath string, data []byte, perm os.FileMode) error {
	fullPath := filepath.Join(d.path(), relativePath)

	// Create parent directories
	dir := filepath.Dir(fullPath)
	if err := d.fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write the file
	if err := d.fs.WriteFile(fullPath, data, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", relativePath, err)
	}

	return nil
}

// WriteFiles writes multiple files to the destination.
func (d *DestinationImpl) WriteFiles(files map[string][]byte) error {
	for path, data := range files {
		if err := d.WriteFile(path, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// CopyFrom copies all files from the source directory to the destination.
func (d *DestinationImpl) CopyFrom(srcPath string) error {
	files, err := d.fs.ListFiles(srcPath)
	if err != nil {
		return fmt.Errorf("failed to list files in %s: %w", srcPath, err)
	}

	for _, file := range files {
		src := filepath.Join(srcPath, file)
		dst := filepath.Join(d.path(), file)

		// Create parent directory
		if err := d.fs.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file, err)
		}

		// Read source file
		data, err := d.fs.ReadFile(src)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Get source permissions
		srcInfo, err := d.fs.Stat(src)
		perm := os.FileMode(0o644)
		if err == nil {
			perm = srcInfo.Mode().Perm()
		}

		// Write destination file
		if err := d.fs.WriteFile(dst, data, perm); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}
	}

	return nil
}

// Clear removes all files from the destination directory.
func (d *DestinationImpl) Clear() error {
	path := d.path()

	if !d.fs.Exists(path) {
		return nil
	}

	files, err := d.fs.ListFiles(path)
	if err != nil {
		return fmt.Errorf("failed to list files in %s: %w", path, err)
	}

	for _, file := range files {
		fullPath := filepath.Join(path, file)
		if err := d.fs.Remove(fullPath); err != nil {
			return fmt.Errorf("failed to remove %s: %w", file, err)
		}
	}

	return nil
}

// Exists returns true if a file exists in the destination.
func (d *DestinationImpl) Exists(relativePath string) bool {
	fullPath := filepath.Join(d.path(), relativePath)
	return d.fs.Exists(fullPath)
}

// ReadFile reads a file from the destination.
func (d *DestinationImpl) ReadFile(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(d.path(), relativePath)
	return d.fs.ReadFile(fullPath)
}

// ListFiles returns all files in the destination directory.
func (d *DestinationImpl) ListFiles() ([]string, error) {
	path := d.path()

	if !d.fs.Exists(path) {
		return nil, nil
	}

	return d.fs.ListFiles(path)
}

// EnsureCleanDestination ensures the destination exists and is empty.
func (d *DestinationImpl) EnsureCleanDestination() error {
	path := d.path()

	// Create if doesn't exist
	if !d.fs.Exists(path) {
		return d.fs.MkdirAll(path, 0o755)
	}

	// Clear if exists
	return d.Clear()
}

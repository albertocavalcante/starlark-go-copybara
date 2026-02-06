package folder

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/albertocavalcante/starlark-go-copybara/vcs"
)

// Compile-time interface verification.
var _ vcs.Origin = (*OriginImpl)(nil)

// OriginImpl implements the vcs.Origin interface for local folders.
type OriginImpl struct {
	origin                     *Origin
	fs                         FileSystem
	materializeOutsideSymlinks bool
}

// NewOriginImpl creates a new OriginImpl from an Origin configuration.
func NewOriginImpl(origin *Origin) *OriginImpl {
	return &OriginImpl{
		origin:                     origin,
		fs:                         NewOSFileSystem(),
		materializeOutsideSymlinks: origin.materializeOutsideSymlinks,
	}
}

// WithFileSystem sets a custom filesystem (useful for testing).
func (o *OriginImpl) WithFileSystem(fs FileSystem) *OriginImpl {
	o.fs = fs
	return o
}

// URL returns the folder path as the URL.
func (o *OriginImpl) URL() string {
	return o.path()
}

// path returns the resolved path, defaulting to current working directory.
func (o *OriginImpl) path() string {
	if o.origin.path != "" {
		return o.origin.path
	}
	// Default to current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// Ref returns a pseudo-ref based on current time.
func (o *OriginImpl) Ref() string {
	return fmt.Sprintf("folder-%d", time.Now().Unix())
}

// Checkout checks out the given reference (no-op for folders).
func (o *OriginImpl) Checkout(ref string) error {
	// For folder origin, checkout is a no-op since files are already local
	return nil
}

// Changes returns pseudo-changes for the folder contents.
// For folder origins, we generate a single change containing all files.
func (o *OriginImpl) Changes(baseline string) ([]*vcs.Change, error) {
	path := o.path()

	files, err := o.ListFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in %s: %w", path, err)
	}

	// Generate a reference based on file contents
	ref := o.generateRef(files)

	change := &vcs.Change{
		Ref:     ref,
		Author:  "folder.origin",
		Message: fmt.Sprintf("Files from folder: %s", path),
		Files:   files,
	}

	return []*vcs.Change{change}, nil
}

// ListFiles returns all files in the origin directory.
func (o *OriginImpl) ListFiles() ([]string, error) {
	path := o.path()

	if !o.fs.Exists(path) {
		return nil, fmt.Errorf("origin path does not exist: %s", path)
	}

	if !o.fs.IsDir(path) {
		return nil, fmt.Errorf("origin path is not a directory: %s", path)
	}

	files, err := o.fs.ListFiles(path)
	if err != nil {
		return nil, err
	}

	// Handle symlinks if configured
	if o.materializeOutsideSymlinks {
		files, err = o.resolveSymlinks(files)
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// ReadFile reads a file from the origin directory.
func (o *OriginImpl) ReadFile(relativePath string) ([]byte, error) {
	fullPath := filepath.Join(o.path(), relativePath)

	// Handle symlinks if configured
	if o.materializeOutsideSymlinks && o.fs.IsSymlink(fullPath) {
		target, err := o.fs.ReadLink(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read symlink %s: %w", relativePath, err)
		}

		// Resolve relative symlinks
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(fullPath), target)
		}

		// Check if target is outside the origin directory
		originPath := o.path()
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return nil, err
		}
		originAbs, err := filepath.Abs(originPath)
		if err != nil {
			return nil, err
		}

		rel, err := filepath.Rel(originAbs, targetAbs)
		if err != nil || len(rel) > 2 && rel[:2] == ".." {
			// Target is outside origin, read the actual file
			return o.fs.ReadFile(target)
		}
	}

	return o.fs.ReadFile(fullPath)
}

// resolveSymlinks handles symlink resolution for file listing.
func (o *OriginImpl) resolveSymlinks(files []string) ([]string, error) {
	// For now, just return the files as-is
	// Full symlink resolution would walk symlinks and include their targets
	return files, nil
}

// generateRef generates a reference hash based on file list.
func (o *OriginImpl) generateRef(files []string) string {
	h := sha256.New()
	for _, f := range files {
		h.Write([]byte(f))
	}
	return "folder-" + hex.EncodeToString(h.Sum(nil))[:12]
}

// CopyTo copies all files from the origin to the destination directory.
func (o *OriginImpl) CopyTo(destPath string) error {
	files, err := o.ListFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(o.path(), file)
		dstPath := filepath.Join(destPath, file)

		// Create parent directory
		if err := o.fs.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file, err)
		}

		// Read source file
		data, err := o.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", file, err)
		}

		// Get source permissions
		srcInfo, err := o.fs.Stat(srcPath)
		perm := os.FileMode(0644)
		if err == nil {
			perm = srcInfo.Mode().Perm()
		}

		// Write destination file
		if err := o.fs.WriteFile(dstPath, data, perm); err != nil {
			return fmt.Errorf("failed to write %s: %w", file, err)
		}
	}

	return nil
}

// FileInfo represents information about a file in the origin.
type FileInfo struct {
	Path    string
	Size    int64
	Mode    os.FileMode
	IsDir   bool
	ModTime time.Time
}

// Stat returns information about a file in the origin.
func (o *OriginImpl) Stat(relativePath string) (*FileInfo, error) {
	fullPath := filepath.Join(o.path(), relativePath)
	info, err := o.fs.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Path:    relativePath,
		Size:    info.Size(),
		Mode:    info.Mode(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
	}, nil
}

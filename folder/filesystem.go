// Package folder provides the folder.* Starlark module for local folder operations.
package folder

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileSystem is an abstraction for file operations.
// It supports both real filesystem and in-memory filesystem (for WASM).
type FileSystem interface {
	// ReadFile reads the entire contents of a file.
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file, creating it if necessary.
	WriteFile(path string, data []byte, perm fs.FileMode) error

	// ListFiles returns all files in a directory recursively.
	ListFiles(dir string) ([]string, error)

	// Exists returns true if the path exists.
	Exists(path string) bool

	// IsDir returns true if the path is a directory.
	IsDir(path string) bool

	// MkdirAll creates a directory and all parent directories.
	MkdirAll(path string, perm fs.FileMode) error

	// Remove removes a file or empty directory.
	Remove(path string) error

	// RemoveAll removes a path and all its children.
	RemoveAll(path string) error

	// Stat returns file info for the given path.
	Stat(path string) (fs.FileInfo, error)

	// ReadLink returns the destination of a symbolic link.
	ReadLink(path string) (string, error)

	// IsSymlink returns true if the path is a symbolic link.
	IsSymlink(path string) bool
}

// OSFileSystem implements FileSystem using the real operating system filesystem.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OSFileSystem.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile reads the entire contents of a file.
func (f *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path) //nolint:gosec // path comes from trusted callers
}

// WriteFile writes data to a file, creating it if necessary.
func (f *OSFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// ListFiles returns all files in a directory recursively.
func (f *OSFileSystem) ListFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			// Return relative path from dir
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

// Exists returns true if the path exists.
func (f *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir returns true if the path is a directory.
func (f *OSFileSystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// MkdirAll creates a directory and all parent directories.
func (f *OSFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Remove removes a file or empty directory.
func (f *OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll removes a path and all its children.
func (f *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Stat returns file info for the given path.
func (f *OSFileSystem) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// ReadLink returns the destination of a symbolic link.
func (f *OSFileSystem) ReadLink(path string) (string, error) {
	return os.Readlink(path)
}

// IsSymlink returns true if the path is a symbolic link.
func (f *OSFileSystem) IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// MemoryFileSystem implements FileSystem using an in-memory map.
// This is useful for WASM environments where real filesystem access is limited.
type MemoryFileSystem struct {
	mu    sync.RWMutex
	files map[string]*memFile
}

type memFile struct {
	data  []byte
	perm  fs.FileMode
	isDir bool
}

// NewMemoryFileSystem creates a new in-memory filesystem.
func NewMemoryFileSystem() *MemoryFileSystem {
	return &MemoryFileSystem{
		files: make(map[string]*memFile),
	}
}

// ReadFile reads the entire contents of a file.
func (f *MemoryFileSystem) ReadFile(path string) ([]byte, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	path = filepath.Clean(path)
	file, ok := f.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if file.isDir {
		return nil, &fs.PathError{Op: "read", Path: path, Err: fs.ErrInvalid}
	}
	// Return a copy to prevent mutation
	data := make([]byte, len(file.data))
	copy(data, file.data)
	return data, nil
}

// WriteFile writes data to a file, creating it if necessary.
func (f *MemoryFileSystem) WriteFile(path string, data []byte, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	path = filepath.Clean(path)

	// Create parent directories
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		_ = f.mkdirAllLocked(dir, 0o755)
	}

	// Copy data to prevent mutation
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	f.files[path] = &memFile{
		data:  dataCopy,
		perm:  perm,
		isDir: false,
	}
	return nil
}

// ListFiles returns all files in a directory recursively.
func (f *MemoryFileSystem) ListFiles(dir string) ([]string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	dir = filepath.Clean(dir)
	var files []string

	for path, file := range f.files {
		if file.isDir {
			continue
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			continue
		}
		// Only include files within the directory (not starting with ..)
		if rel != "" && rel[0] != '.' {
			files = append(files, rel)
		} else if rel == filepath.Base(path) && filepath.Dir(path) == dir {
			files = append(files, rel)
		}
	}
	return files, nil
}

// Exists returns true if the path exists.
func (f *MemoryFileSystem) Exists(path string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	path = filepath.Clean(path)
	_, ok := f.files[path]
	return ok
}

// IsDir returns true if the path is a directory.
func (f *MemoryFileSystem) IsDir(path string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	path = filepath.Clean(path)
	file, ok := f.files[path]
	return ok && file.isDir
}

// MkdirAll creates a directory and all parent directories.
func (f *MemoryFileSystem) MkdirAll(path string, perm fs.FileMode) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.mkdirAllLocked(path, perm)
}

func (f *MemoryFileSystem) mkdirAllLocked(path string, perm fs.FileMode) error {
	path = filepath.Clean(path)
	if path == "." || path == "/" {
		return nil
	}

	// Create parent first
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" {
		if err := f.mkdirAllLocked(parent, perm); err != nil {
			return err
		}
	}

	// Create this directory if it doesn't exist
	if existing, ok := f.files[path]; ok {
		if !existing.isDir {
			return &fs.PathError{Op: "mkdir", Path: path, Err: fs.ErrExist}
		}
		return nil
	}

	f.files[path] = &memFile{
		perm:  perm,
		isDir: true,
	}
	return nil
}

// Remove removes a file or empty directory.
func (f *MemoryFileSystem) Remove(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	path = filepath.Clean(path)
	if _, ok := f.files[path]; !ok {
		return fs.ErrNotExist
	}
	delete(f.files, path)
	return nil
}

// RemoveAll removes a path and all its children.
func (f *MemoryFileSystem) RemoveAll(path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	path = filepath.Clean(path)
	toDelete := []string{}
	for p := range f.files {
		if p == path || hasPrefix(p, path+string(filepath.Separator)) {
			toDelete = append(toDelete, p)
		}
	}
	for _, p := range toDelete {
		delete(f.files, p)
	}
	return nil
}

// Stat returns file info for the given path.
func (f *MemoryFileSystem) Stat(path string) (fs.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	path = filepath.Clean(path)
	file, ok := f.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return &memFileInfo{
		name:  filepath.Base(path),
		size:  int64(len(file.data)),
		mode:  file.perm,
		isDir: file.isDir,
	}, nil
}

// ReadLink returns the destination of a symbolic link.
// Memory filesystem does not support symlinks.
func (f *MemoryFileSystem) ReadLink(path string) (string, error) {
	return "", &fs.PathError{Op: "readlink", Path: path, Err: fs.ErrInvalid}
}

// IsSymlink returns true if the path is a symbolic link.
// Memory filesystem does not support symlinks.
func (f *MemoryFileSystem) IsSymlink(path string) bool {
	return false
}

// memFileInfo implements fs.FileInfo for memory files.
type memFileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	isDir   bool
	modTime time.Time
}

func (fi *memFileInfo) Name() string       { return fi.name }
func (fi *memFileInfo) Size() int64        { return fi.size }
func (fi *memFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *memFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *memFileInfo) IsDir() bool        { return fi.isDir }
func (fi *memFileInfo) Sys() any           { return nil }

// hasPrefix checks if path has the given prefix.
func hasPrefix(path, prefix string) bool {
	return len(path) >= len(prefix) && path[:len(prefix)] == prefix
}

// CopyFile copies a file from src to dst.
func CopyFile(fsys FileSystem, src, dst string) error {
	data, err := fsys.ReadFile(src)
	if err != nil {
		return err
	}

	// Get source permissions if possible
	perm := os.FileMode(0o644)
	if info, err := fsys.Stat(src); err == nil {
		perm = info.Mode().Perm()
	}

	return fsys.WriteFile(dst, data, perm)
}

// CopyDir copies a directory recursively from src to dst.
func CopyDir(fsys FileSystem, src, dst string) error {
	files, err := fsys.ListFiles(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file)
		dstPath := filepath.Join(dst, file)

		// Create parent directory
		if err := fsys.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
			return err
		}

		if err := CopyFile(fsys, srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}

// ReadAll reads all content from a reader.
func ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

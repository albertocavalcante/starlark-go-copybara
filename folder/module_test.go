package folder_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/folder"
)

func TestModule(t *testing.T) {
	if folder.Module == nil {
		t.Fatal("expected non-nil module")
	}

	if folder.Module.Name != "folder" {
		t.Errorf("expected module name 'folder', got %q", folder.Module.Name)
	}

	// Verify module members
	members := folder.Module.Members
	if _, ok := members["origin"]; !ok {
		t.Error("expected 'origin' member in module")
	}
	if _, ok := members["destination"]; !ok {
		t.Error("expected 'destination' member in module")
	}
}

func TestOriginWithPath(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `folder.origin(path = "/tmp/source")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	origin, ok := val.(*folder.Origin)
	if !ok {
		t.Fatalf("expected *Origin, got %T", val)
	}

	if origin.Path() != "/tmp/source" {
		t.Errorf("expected path '/tmp/source', got %q", origin.Path())
	}

	expected := `folder.origin(path = "/tmp/source")`
	if origin.String() != expected {
		t.Errorf("expected string %q, got %q", expected, origin.String())
	}
}

func TestOriginWithoutPath(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `folder.origin()`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	origin, ok := val.(*folder.Origin)
	if !ok {
		t.Fatalf("expected *Origin, got %T", val)
	}

	// Path should default to current working directory
	wd, _ := os.Getwd()
	if origin.Path() != wd {
		t.Errorf("expected path %q, got %q", wd, origin.Path())
	}

	if origin.String() != "folder.origin()" {
		t.Errorf("expected string 'folder.origin()', got %q", origin.String())
	}
}

func TestOriginWithMaterializeSymlinks(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`folder.origin(path = "/tmp/src", materialize_outside_symlinks = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	origin, ok := val.(*folder.Origin)
	if !ok {
		t.Fatalf("expected *Origin, got %T", val)
	}

	if !origin.MaterializeOutsideSymlinks() {
		t.Error("expected materialize_outside_symlinks to be true")
	}
}

func TestOriginAttrs(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	// Test attribute access
	val, err := starlark.Eval(thread, "test.sky",
		`folder.origin(path = "/test/path").path`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathVal, ok := val.(starlark.String)
	if !ok {
		t.Fatalf("expected starlark.String, got %T", val)
	}

	if string(pathVal) != "/test/path" {
		t.Errorf("expected path '/test/path', got %q", pathVal)
	}
}

func TestDestinationWithPath(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `folder.destination(path = "/tmp/dest")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest, ok := val.(*folder.Destination)
	if !ok {
		t.Fatalf("expected *Destination, got %T", val)
	}

	if dest.Path() != "/tmp/dest" {
		t.Errorf("expected path '/tmp/dest', got %q", dest.Path())
	}

	expected := `folder.destination(path = "/tmp/dest")`
	if dest.String() != expected {
		t.Errorf("expected string %q, got %q", expected, dest.String())
	}
}

func TestDestinationWithoutPath(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `folder.destination()`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest, ok := val.(*folder.Destination)
	if !ok {
		t.Fatalf("expected *Destination, got %T", val)
	}

	// Path should default to current working directory
	wd, _ := os.Getwd()
	if dest.Path() != wd {
		t.Errorf("expected path %q, got %q", wd, dest.Path())
	}
}

func TestDestinationAttrs(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	// Test attribute access
	val, err := starlark.Eval(thread, "test.sky",
		`folder.destination(path = "/test/dest").path`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pathVal, ok := val.(starlark.String)
	if !ok {
		t.Fatalf("expected starlark.String, got %T", val)
	}

	if string(pathVal) != "/test/dest" {
		t.Errorf("expected path '/test/dest', got %q", pathVal)
	}
}

func TestOriginReadFiles(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]string{
		"file1.txt":        "content of file 1",
		"file2.txt":        "content of file 2",
		"subdir/file3.txt": "content of file 3",
	}

	for path, content := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file: %v", err)
		}
	}

	// Create origin
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	script := `folder.origin(path = "` + tmpDir + `")`
	val, err := starlark.Eval(thread, "test.sky", script, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	origin := val.(*folder.Origin)
	impl := origin.Impl()

	// Test listing files
	files, err := impl.ListFiles()
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d: %v", len(files), files)
	}

	// Test reading files
	for path, expectedContent := range testFiles {
		content, err := impl.ReadFile(path)
		if err != nil {
			t.Errorf("failed to read file %s: %v", path, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("file %s: expected content %q, got %q", path, expectedContent, string(content))
		}
	}
}

func TestDestinationWriteFiles(t *testing.T) {
	// Create a temporary directory for destination
	tmpDir := t.TempDir()

	// Create destination
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	script := `folder.destination(path = "` + tmpDir + `")`
	val, err := starlark.Eval(thread, "test.sky", script, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dest := val.(*folder.Destination)
	impl := dest.Impl()

	// Test writing files
	testFiles := map[string][]byte{
		"output1.txt":        []byte("output content 1"),
		"output2.txt":        []byte("output content 2"),
		"subdir/output3.txt": []byte("output content 3"),
	}

	if err := impl.WriteFiles(testFiles); err != nil {
		t.Fatalf("failed to write files: %v", err)
	}

	// Verify files were written
	for path, expectedContent := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read written file %s: %v", path, err)
			continue
		}
		if string(content) != string(expectedContent) {
			t.Errorf("file %s: expected content %q, got %q", path, expectedContent, string(content))
		}
	}
}

func TestFullWorkflow(t *testing.T) {
	// Create source directory with files
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files
	sourceFiles := map[string]string{
		"src/main.go":    "package main\n\nfunc main() {}\n",
		"src/util.go":    "package main\n\nfunc helper() {}\n",
		"README.md":      "# Test Project\n",
		"config/app.yml": "name: test\nversion: 1.0\n",
	}

	for path, content := range sourceFiles {
		fullPath := filepath.Join(srcDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write source file: %v", err)
		}
	}

	// Create origin and destination
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"folder": folder.Module,
	}

	originScript := `folder.origin(path = "` + srcDir + `")`
	originVal, err := starlark.Eval(thread, "test.sky", originScript, predeclared)
	if err != nil {
		t.Fatalf("failed to create origin: %v", err)
	}

	destScript := `folder.destination(path = "` + dstDir + `")`
	destVal, err := starlark.Eval(thread, "test.sky", destScript, predeclared)
	if err != nil {
		t.Fatalf("failed to create destination: %v", err)
	}

	origin := originVal.(*folder.Origin)
	dest := destVal.(*folder.Destination)

	originImpl := origin.Impl()
	destImpl := dest.Impl()

	// Copy files from origin to destination
	files, err := originImpl.ListFiles()
	if err != nil {
		t.Fatalf("failed to list origin files: %v", err)
	}

	for _, file := range files {
		content, err := originImpl.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", file, err)
		}

		if err := destImpl.WriteFile(file, content, 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", file, err)
		}
	}

	// Verify all files were copied
	for path, expectedContent := range sourceFiles {
		fullPath := filepath.Join(dstDir, path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Errorf("failed to read destination file %s: %v", path, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("file %s: expected content %q, got %q", path, expectedContent, string(content))
		}
	}
}

func TestOriginCopyTo(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files
	testContent := "test file content"
	testPath := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testPath, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Create origin and copy
	// Use the module to create it properly
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "`+srcDir+`")`, predeclared)
	origin := val.(*folder.Origin)

	impl := origin.Impl()
	if err := impl.CopyTo(dstDir); err != nil {
		t.Fatalf("failed to copy: %v", err)
	}

	// Verify
	content, err := os.ReadFile(filepath.Join(dstDir, "test.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("expected %q, got %q", testContent, string(content))
	}
}

func TestDestinationCopyFrom(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files
	testContent := "source content"
	if err := os.WriteFile(filepath.Join(srcDir, "source.txt"), []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Create destination and copy
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "`+dstDir+`")`, predeclared)
	dest := val.(*folder.Destination)

	impl := dest.Impl()
	if err := impl.CopyFrom(srcDir); err != nil {
		t.Fatalf("failed to copy from source: %v", err)
	}

	// Verify
	content, err := os.ReadFile(filepath.Join(dstDir, "source.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != testContent {
		t.Errorf("expected %q, got %q", testContent, string(content))
	}
}

func TestDestinationClear(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some files
	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "`+tmpDir+`")`, predeclared)
	dest := val.(*folder.Destination)

	impl := dest.Impl()
	if err := impl.Clear(); err != nil {
		t.Fatalf("failed to clear: %v", err)
	}

	// Verify files are removed
	files, err := impl.ListFiles()
	if err != nil {
		t.Fatalf("failed to list files: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty directory, got %d files: %v", len(files), files)
	}
}

func TestOriginChanges(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "`+tmpDir+`")`, predeclared)
	origin := val.(*folder.Origin)

	impl := origin.Impl()
	changes, err := impl.Changes("")
	if err != nil {
		t.Fatalf("failed to get changes: %v", err)
	}

	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}

	change := changes[0]
	if change.Author != "folder.origin" {
		t.Errorf("expected author 'folder.origin', got %q", change.Author)
	}
	if len(change.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(change.Files))
	}
	if !slices.Contains(change.Files, "file.txt") {
		t.Errorf("expected 'file.txt' in files, got %v", change.Files)
	}
}

func TestMemoryFileSystem(t *testing.T) {
	fs := folder.NewMemoryFileSystem()

	// Test WriteFile and ReadFile
	content := []byte("test content")
	if err := fs.WriteFile("/test/file.txt", content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	read, err := fs.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(read) != string(content) {
		t.Errorf("expected %q, got %q", content, read)
	}

	// Test Exists
	if !fs.Exists("/test/file.txt") {
		t.Error("expected file to exist")
	}
	if fs.Exists("/nonexistent") {
		t.Error("expected file to not exist")
	}

	// Test IsDir
	if !fs.IsDir("/test") {
		t.Error("expected /test to be a directory")
	}
	if fs.IsDir("/test/file.txt") {
		t.Error("expected /test/file.txt to not be a directory")
	}

	// Test Remove
	if err := fs.Remove("/test/file.txt"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if fs.Exists("/test/file.txt") {
		t.Error("expected file to be removed")
	}
}

func TestOriginWithMemoryFileSystem(t *testing.T) {
	// Create an origin with memory filesystem for WASM compatibility testing
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "/virtual")`, predeclared)
	origin := val.(*folder.Origin)

	memFS := folder.NewMemoryFileSystem()
	memFS.WriteFile("/virtual/test.txt", []byte("virtual content"), 0644)

	impl := origin.Impl().WithFileSystem(memFS)

	// Test reading from memory filesystem
	content, err := impl.ReadFile("test.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "virtual content" {
		t.Errorf("expected 'virtual content', got %q", string(content))
	}
}

func TestDestinationWithMemoryFileSystem(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "/virtual/output")`, predeclared)
	dest := val.(*folder.Destination)

	memFS := folder.NewMemoryFileSystem()
	impl := dest.Impl().WithFileSystem(memFS)

	// Test writing to memory filesystem
	if err := impl.WriteFile("output.txt", []byte("output content"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Verify
	content, err := memFS.ReadFile("/virtual/output/output.txt")
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if string(content) != "output content" {
		t.Errorf("expected 'output content', got %q", string(content))
	}
}

func TestOriginStarlarkValue(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "/test")`, predeclared)
	origin := val.(*folder.Origin)

	// Test Type
	if origin.Type() != "folder.origin" {
		t.Errorf("expected type 'folder.origin', got %q", origin.Type())
	}

	// Test Truth
	if origin.Truth() != starlark.True {
		t.Error("expected Truth to return True")
	}

	// Test Hash (should error)
	_, err := origin.Hash()
	if err == nil {
		t.Error("expected Hash to return error")
	}

	// Test Freeze (should not panic)
	origin.Freeze()
}

func TestDestinationStarlarkValue(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "/test")`, predeclared)
	dest := val.(*folder.Destination)

	// Test Type
	if dest.Type() != "folder.destination" {
		t.Errorf("expected type 'folder.destination', got %q", dest.Type())
	}

	// Test Truth
	if dest.Truth() != starlark.True {
		t.Error("expected Truth to return True")
	}

	// Test Hash (should error)
	_, err := dest.Hash()
	if err == nil {
		t.Error("expected Hash to return error")
	}

	// Test Freeze (should not panic)
	dest.Freeze()
}

func TestOriginNonExistentPath(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "/nonexistent/path")`, predeclared)
	origin := val.(*folder.Origin)
	impl := origin.Impl()

	_, err := impl.ListFiles()
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestDestinationEnsureClean(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "new_dest")

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "`+destPath+`")`, predeclared)
	dest := val.(*folder.Destination)
	impl := dest.Impl()

	// Should create directory if it doesn't exist
	if err := impl.EnsureCleanDestination(); err != nil {
		t.Fatalf("EnsureCleanDestination failed: %v", err)
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("expected destination directory to be created")
	}
}

func TestOSFileSystem(t *testing.T) {
	tmpDir := t.TempDir()
	fs := folder.NewOSFileSystem()

	// Test WriteFile and ReadFile
	testPath := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	if err := fs.WriteFile(testPath, content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	read, err := fs.ReadFile(testPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(read) != string(content) {
		t.Errorf("expected %q, got %q", content, read)
	}

	// Test Exists
	if !fs.Exists(testPath) {
		t.Error("expected file to exist")
	}

	// Test Stat
	info, err := fs.Stat(testPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), info.Size())
	}

	// Test IsDir
	if fs.IsDir(testPath) {
		t.Error("expected file to not be a directory")
	}
	if !fs.IsDir(tmpDir) {
		t.Error("expected tmpDir to be a directory")
	}

	// Test MkdirAll
	subDir := filepath.Join(tmpDir, "sub", "dir")
	if err := fs.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if !fs.IsDir(subDir) {
		t.Error("expected subDir to be created")
	}

	// Test ListFiles
	files, err := fs.ListFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d: %v", len(files), files)
	}

	// Test Remove
	if err := fs.Remove(testPath); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if fs.Exists(testPath) {
		t.Error("expected file to be removed")
	}

	// Test RemoveAll
	if err := fs.RemoveAll(subDir); err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}
	if fs.Exists(subDir) {
		t.Error("expected directory to be removed")
	}
}

func TestOSFileSystemSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	fs := folder.NewOSFileSystem()

	// Create a file
	targetPath := filepath.Join(tmpDir, "target.txt")
	if err := os.WriteFile(targetPath, []byte("target content"), 0644); err != nil {
		t.Fatalf("failed to write target file: %v", err)
	}

	// Create symlink
	linkPath := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	// Test IsSymlink
	if !fs.IsSymlink(linkPath) {
		t.Error("expected link.txt to be a symlink")
	}
	if fs.IsSymlink(targetPath) {
		t.Error("expected target.txt to not be a symlink")
	}

	// Test ReadLink
	target, err := fs.ReadLink(linkPath)
	if err != nil {
		t.Fatalf("ReadLink failed: %v", err)
	}
	if target != targetPath {
		t.Errorf("expected target %q, got %q", targetPath, target)
	}
}

func TestMemoryFileSystemRemoveAll(t *testing.T) {
	fs := folder.NewMemoryFileSystem()

	// Create nested structure
	fs.WriteFile("/root/dir1/file1.txt", []byte("content1"), 0644)
	fs.WriteFile("/root/dir1/file2.txt", []byte("content2"), 0644)
	fs.WriteFile("/root/dir2/file3.txt", []byte("content3"), 0644)

	// Remove dir1
	if err := fs.RemoveAll("/root/dir1"); err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify dir1 files are gone
	if fs.Exists("/root/dir1/file1.txt") {
		t.Error("expected file1.txt to be removed")
	}
	if fs.Exists("/root/dir1/file2.txt") {
		t.Error("expected file2.txt to be removed")
	}

	// Verify dir2 files still exist
	if !fs.Exists("/root/dir2/file3.txt") {
		t.Error("expected file3.txt to still exist")
	}
}

func TestMemoryFileSystemStat(t *testing.T) {
	fs := folder.NewMemoryFileSystem()

	content := []byte("test content")
	fs.WriteFile("/test.txt", content, 0644)

	info, err := fs.Stat("/test.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Name() != "test.txt" {
		t.Errorf("expected name 'test.txt', got %q", info.Name())
	}
	if info.Size() != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), info.Size())
	}
	if info.IsDir() {
		t.Error("expected IsDir to be false")
	}

	// Test stat on non-existent file
	_, err = fs.Stat("/nonexistent")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestMemoryFileSystemSymlinks(t *testing.T) {
	fs := folder.NewMemoryFileSystem()

	// Memory filesystem doesn't support symlinks
	if fs.IsSymlink("/any/path") {
		t.Error("memory filesystem should never have symlinks")
	}

	_, err := fs.ReadLink("/any/path")
	if err == nil {
		t.Error("expected error from ReadLink on memory filesystem")
	}
}

func TestOriginStat(t *testing.T) {
	tmpDir := t.TempDir()
	content := []byte("test content for stat")
	testPath := filepath.Join(tmpDir, "stat_test.txt")
	if err := os.WriteFile(testPath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "`+tmpDir+`")`, predeclared)
	origin := val.(*folder.Origin)
	impl := origin.Impl()

	info, err := impl.Stat("stat_test.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Path != "stat_test.txt" {
		t.Errorf("expected path 'stat_test.txt', got %q", info.Path)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), info.Size)
	}
	if info.IsDir {
		t.Error("expected IsDir to be false")
	}
}

func TestOriginURL(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "/test/path")`, predeclared)
	origin := val.(*folder.Origin)
	impl := origin.Impl()

	if impl.URL() != "/test/path" {
		t.Errorf("expected URL '/test/path', got %q", impl.URL())
	}
}

func TestOriginCheckout(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.origin(path = "/test")`, predeclared)
	origin := val.(*folder.Origin)
	impl := origin.Impl()

	// Checkout should be a no-op
	if err := impl.Checkout("any-ref"); err != nil {
		t.Errorf("Checkout should succeed: %v", err)
	}
}

func TestDestinationURL(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "/dest/path")`, predeclared)
	dest := val.(*folder.Destination)
	impl := dest.Impl()

	if impl.URL() != "/dest/path" {
		t.Errorf("expected URL '/dest/path', got %q", impl.URL())
	}
}

func TestDestinationCheckout(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}

	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "/test")`, predeclared)
	dest := val.(*folder.Destination)
	impl := dest.Impl()

	// Checkout should be a no-op
	if err := impl.Checkout("any-ref"); err != nil {
		t.Errorf("Checkout should succeed: %v", err)
	}

	// Ref should return empty string
	if impl.Ref() != "" {
		t.Errorf("expected empty Ref, got %q", impl.Ref())
	}
}

func TestDestinationExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	if err := os.WriteFile(filepath.Join(tmpDir, "exists.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "`+tmpDir+`")`, predeclared)
	dest := val.(*folder.Destination)
	impl := dest.Impl()

	if !impl.Exists("exists.txt") {
		t.Error("expected file to exist")
	}
	if impl.Exists("nonexistent.txt") {
		t.Error("expected file to not exist")
	}
}

func TestDestinationReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	content := "test content"

	if err := os.WriteFile(filepath.Join(tmpDir, "read.txt"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{"folder": folder.Module}
	val, _ := starlark.Eval(thread, "test.sky", `folder.destination(path = "`+tmpDir+`")`, predeclared)
	dest := val.(*folder.Destination)
	impl := dest.Impl()

	read, err := impl.ReadFile("read.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(read) != content {
		t.Errorf("expected %q, got %q", content, string(read))
	}
}

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create source files
	if err := os.MkdirAll(filepath.Join(srcDir, "sub"), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "sub", "file2.txt"), []byte("content2"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	fs := folder.NewOSFileSystem()
	if err := folder.CopyDir(fs, srcDir, dstDir); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// Verify
	content1, err := os.ReadFile(filepath.Join(dstDir, "file1.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content1) != "content1" {
		t.Errorf("expected 'content1', got %q", string(content1))
	}

	content2, err := os.ReadFile(filepath.Join(dstDir, "sub", "file2.txt"))
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content2) != "content2" {
		t.Errorf("expected 'content2', got %q", string(content2))
	}
}

package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

func TestCopyCreation(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "basic copy",
			code:    `core.copy("src", "dst")`,
			wantErr: false,
		},
		{
			name:    "copy with paths",
			code:    `core.copy("src", "dst", paths = ["*.go"])`,
			wantErr: false,
		},
		{
			name:    "copy with glob",
			code:    `core.copy("src", "dst", paths = core.glob(include = ["*.go"]))`,
			wantErr: false,
		},
		{
			name:    "copy with overwrite",
			code:    `core.copy("src", "dst", overwrite = True)`,
			wantErr: false,
		},
		{
			name:    "copy same path",
			code:    `core.copy("src", "src")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			copy, ok := val.(*core.Copy)
			if !ok {
				t.Fatalf("expected *Copy, got %T", val)
			}

			if copy.Type() != "copy" {
				t.Errorf("Type() = %q, want %q", copy.Type(), "copy")
			}
		})
	}
}

func TestCopyString(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.copy("src", "dst")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copy := val.(*core.Copy)
	expected := `core.copy("src", "dst")`
	if got := copy.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}
}

func TestCopyApply(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "copy_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source directory and file
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	srcFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(srcFile, []byte("hello world"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Create and apply copy transformation
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.copy("src", "dst")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copy := val.(*core.Copy)
	ctx := &transform.Context{WorkDir: tmpDir}

	if err := copy.Apply(ctx); err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify destination file exists
	dstFile := filepath.Join(tmpDir, "dst", "test.txt")
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("destination content = %q, want %q", string(content), "hello world")
	}

	// Verify source file still exists (it's a copy, not move)
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("source file should still exist after copy")
	}
}

func TestCopyWithGlob(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "copy_glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source directory with multiple files
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	// Create .go and .txt files
	if err := os.WriteFile(filepath.Join(srcDir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "readme.txt"), []byte("readme"), 0o644); err != nil {
		t.Fatalf("failed to write readme.txt: %v", err)
	}

	// Create and apply copy transformation with glob
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.copy("src", "dst", paths = ["*.go"])`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copy := val.(*core.Copy)
	ctx := &transform.Context{WorkDir: tmpDir}

	if err := copy.Apply(ctx); err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify .go file was copied
	if _, err := os.Stat(filepath.Join(tmpDir, "dst", "main.go")); os.IsNotExist(err) {
		t.Error("main.go should have been copied")
	}

	// Verify .txt file was NOT copied
	if _, err := os.Stat(filepath.Join(tmpDir, "dst", "readme.txt")); !os.IsNotExist(err) {
		t.Error("readme.txt should NOT have been copied")
	}
}

func TestCopyReverse(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.copy("src", "dst")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copy := val.(*core.Copy)
	reverse := copy.Reverse()

	// Reverse of copy should be remove
	if reverse.Describe() != "Removing files matching glob(include = [\"dst\", \"dst/**\"])" {
		t.Errorf("Reverse().Describe() = %q, unexpected", reverse.Describe())
	}
}

func TestCopyDescribe(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.copy("src", "dst")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	copy := val.(*core.Copy)
	expected := "Copying src to dst"
	if got := copy.Describe(); got != expected {
		t.Errorf("Describe() = %q, want %q", got, expected)
	}
}

func TestCopyOverwrite(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "copy_overwrite_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "src.txt")
	if err := os.WriteFile(srcFile, []byte("new content"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// Create existing destination file
	dstFile := filepath.Join(tmpDir, "dst.txt")
	if err := os.WriteFile(dstFile, []byte("old content"), 0o644); err != nil {
		t.Fatalf("failed to write destination file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	// Test without overwrite (should fail)
	val, _ := starlark.Eval(thread, "test.sky", `core.copy("src.txt", "dst.txt")`, predeclared)
	copy := val.(*core.Copy)
	ctx := &transform.Context{WorkDir: tmpDir}

	if err := copy.Apply(ctx); err == nil {
		t.Error("expected error when destination exists without overwrite")
	}

	// Test with overwrite (should succeed)
	val, _ = starlark.Eval(thread, "test.sky", `core.copy("src.txt", "dst.txt", overwrite = True)`, predeclared)
	copy = val.(*core.Copy)

	if err := copy.Apply(ctx); err != nil {
		t.Fatalf("Apply() with overwrite failed: %v", err)
	}

	// Verify content was overwritten
	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(content) != "new content" {
		t.Errorf("destination content = %q, want %q", string(content), "new content")
	}
}

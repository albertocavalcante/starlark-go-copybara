package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

func TestModule(t *testing.T) {
	if core.Module == nil {
		t.Fatal("expected non-nil module")
	}

	if core.Module.Name != "core" {
		t.Errorf("expected module name 'core', got %q", core.Module.Name)
	}

	// Check all expected members exist
	expectedMembers := []string{
		"workflow",
		"move",
		"copy",
		"replace",
		"remove",
		"verify_match",
		"glob",
	}

	for _, name := range expectedMembers {
		if _, ok := core.Module.Members[name]; !ok {
			t.Errorf("expected member %q not found in module", name)
		}
	}
}

func TestGlobals(t *testing.T) {
	globals := core.Globals()
	if globals == nil {
		t.Fatal("expected non-nil globals")
	}

	if _, ok := globals["glob"]; !ok {
		t.Error("expected glob function in globals")
	}
}

func TestMove(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.move("src", "lib")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	move, ok := val.(*core.Move)
	if !ok {
		t.Fatalf("expected *Move, got %T", val)
	}

	if move.String() != `core.move("src", "lib")` {
		t.Errorf("unexpected string: %s", move.String())
	}

	if move.Before() != "src" {
		t.Errorf("Before() = %q, want %q", move.Before(), "src")
	}

	if move.After() != "lib" {
		t.Errorf("After() = %q, want %q", move.After(), "lib")
	}
}

func TestMoveWithPaths(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`core.move("src", "lib", paths = ["*.go"])`,
		predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	move := val.(*core.Move)
	if move.Paths() == nil {
		t.Error("Paths() should not be nil")
	}
}

func TestMoveWithOverwrite(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`core.move("src", "lib", overwrite = True)`,
		predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	move := val.(*core.Move)
	if !move.Overwrite() {
		t.Error("Overwrite() should be true")
	}
}

func TestMoveSamePath(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	_, err := starlark.Eval(thread, "test.sky", `core.move("src", "src")`, predeclared)
	if err == nil {
		t.Error("expected error for moving to same path")
	}
}

func TestMoveApply(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "move_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create source file
	srcFile := filepath.Join(tmpDir, "src.txt")
	if err := os.WriteFile(srcFile, []byte("hello"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, _ := starlark.Eval(thread, "test.sky", `core.move("src.txt", "dst.txt")`, predeclared)
	move := val.(*core.Move)

	ctx := &transform.Context{WorkDir: tmpDir}
	if err := move.Apply(ctx); err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify destination exists
	dstFile := filepath.Join(tmpDir, "dst.txt")
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("destination file should exist")
	}

	// Verify source no longer exists
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("source file should no longer exist")
	}
}

func TestMoveReverse(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, _ := starlark.Eval(thread, "test.sky", `core.move("src", "dst")`, predeclared)
	move := val.(*core.Move)

	reverse := move.Reverse()
	reverseMove, ok := reverse.(*core.Move)
	if !ok {
		t.Fatalf("expected *Move, got %T", reverse)
	}

	if reverseMove.Before() != "dst" {
		t.Errorf("Before() = %q, want %q", reverseMove.Before(), "dst")
	}

	if reverseMove.After() != "src" {
		t.Errorf("After() = %q, want %q", reverseMove.After(), "src")
	}
}

func TestReplace(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.replace("foo", "bar")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	replace, ok := val.(*core.Replace)
	if !ok {
		t.Fatalf("expected *Replace, got %T", val)
	}

	if replace.String() != `core.replace("foo", "bar")` {
		t.Errorf("unexpected string: %s", replace.String())
	}

	if replace.Before() != "foo" {
		t.Errorf("Before() = %q, want %q", replace.Before(), "foo")
	}

	if replace.After() != "bar" {
		t.Errorf("After() = %q, want %q", replace.After(), "bar")
	}
}

func TestReplaceWithPaths(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`core.replace("foo", "bar", paths = ["*.txt"])`,
		predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	replace := val.(*core.Replace)
	if replace.Paths() == nil {
		t.Error("Paths() should not be nil")
	}
}

func TestReplaceApply(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "replace_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("foo bar foo"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, _ := starlark.Eval(thread, "test.sky", `core.replace("foo", "baz")`, predeclared)
	replace := val.(*core.Replace)

	ctx := &transform.Context{WorkDir: tmpDir}
	if err := replace.Apply(ctx); err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify content was replaced
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if string(content) != "baz bar baz" {
		t.Errorf("content = %q, want %q", string(content), "baz bar baz")
	}
}

func TestReplaceReverse(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, _ := starlark.Eval(thread, "test.sky", `core.replace("foo", "bar")`, predeclared)
	replace := val.(*core.Replace)

	reverse := replace.Reverse()
	reverseReplace, ok := reverse.(*core.Replace)
	if !ok {
		t.Fatalf("expected *Replace, got %T", reverse)
	}

	if reverseReplace.Before() != "bar" {
		t.Errorf("Before() = %q, want %q", reverseReplace.Before(), "bar")
	}

	if reverseReplace.After() != "foo" {
		t.Errorf("After() = %q, want %q", reverseReplace.After(), "foo")
	}
}

func TestRemove(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
		"glob": core.Globals()["glob"],
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "remove with glob",
			code:    `core.remove(paths = glob(["*.tmp"]))`,
			wantErr: false,
		},
		{
			name:    "remove with list",
			code:    `core.remove(paths = ["*.tmp", "*.bak"])`,
			wantErr: false,
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

			remove, ok := val.(*core.Remove)
			if !ok {
				t.Fatalf("expected *Remove, got %T", val)
			}

			if remove.Type() != "remove" {
				t.Errorf("Type() = %q, want %q", remove.Type(), "remove")
			}
		})
	}
}

func TestRemoveApply(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "remove_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "delete.tmp"), []byte("delete"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, _ := starlark.Eval(thread, "test.sky", `core.remove(paths = ["*.tmp"])`, predeclared)
	remove := val.(*core.Remove)

	ctx := &transform.Context{WorkDir: tmpDir}
	if err := remove.Apply(ctx); err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify .tmp file was deleted
	if _, err := os.Stat(filepath.Join(tmpDir, "delete.tmp")); !os.IsNotExist(err) {
		t.Error("delete.tmp should have been deleted")
	}

	// Verify .txt file still exists
	if _, err := os.Stat(filepath.Join(tmpDir, "keep.txt")); os.IsNotExist(err) {
		t.Error("keep.txt should still exist")
	}
}

func TestWorkflow(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.workflow(name = "default")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf, ok := val.(*core.Workflow)
	if !ok {
		t.Fatalf("expected *Workflow, got %T", val)
	}

	if wf.Name() != "default" {
		t.Errorf("expected name 'default', got %q", wf.Name())
	}
}

func TestIntegration(t *testing.T) {
	// Test a complete workflow configuration
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
		"glob": core.Globals()["glob"],
	}

	code := `
core.workflow(
	name = "migrate",
	origin_files = glob(["**/*.go"], exclude = ["vendor/**"]),
	destination_files = glob(["**"]),
	transformations = [
		core.move("internal", "pkg"),
		core.replace("old_package", "new_package"),
	],
	mode = "SQUASH",
	reversible_check = True
)
`

	val, err := starlark.Eval(thread, "test.sky", code, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf, ok := val.(*core.Workflow)
	if !ok {
		t.Fatalf("expected *Workflow, got %T", val)
	}

	if wf.Name() != "migrate" {
		t.Errorf("Name() = %q, want %q", wf.Name(), "migrate")
	}

	if wf.Mode() != core.ModeSquash {
		t.Errorf("Mode() = %v, want %v", wf.Mode(), core.ModeSquash)
	}

	if !wf.ReversibleCheck() {
		t.Error("ReversibleCheck() should be true")
	}

	if len(wf.Transformations()) != 2 {
		t.Errorf("len(Transformations()) = %d, want 2", len(wf.Transformations()))
	}

	originFiles := wf.OriginFiles()
	if originFiles == nil {
		t.Fatal("OriginFiles() should not be nil")
	}

	// Test that the glob matches correctly
	if !originFiles.Matches("main.go") {
		t.Error("OriginFiles() should match main.go")
	}

	if originFiles.Matches("vendor/pkg/mod.go") {
		t.Error("OriginFiles() should not match vendor files")
	}
}

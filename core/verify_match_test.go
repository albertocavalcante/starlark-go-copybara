package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

func TestVerifyMatchCreation(t *testing.T) {
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
			name:    "basic verify_match",
			code:    `core.verify_match(regex = "TODO")`,
			wantErr: false,
		},
		{
			name:    "with paths",
			code:    `core.verify_match(regex = "TODO", paths = ["*.go"])`,
			wantErr: false,
		},
		{
			name:    "with verify_no_match",
			code:    `core.verify_match(regex = "FIXME", verify_no_match = True)`,
			wantErr: false,
		},
		{
			name:    "with also_on_reversal",
			code:    `core.verify_match(regex = "TODO", also_on_reversal = True)`,
			wantErr: false,
		},
		{
			name:    "with failure_message",
			code:    `core.verify_match(regex = "TODO", failure_message = "Please add TODO comments")`,
			wantErr: false,
		},
		{
			name:    "invalid regex",
			code:    `core.verify_match(regex = "[invalid")`,
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

			vm, ok := val.(*core.VerifyMatch)
			if !ok {
				t.Fatalf("expected *VerifyMatch, got %T", val)
			}

			if vm.Type() != "verify_match" {
				t.Errorf("Type() = %q, want %q", vm.Type(), "verify_match")
			}
		})
	}
}

func TestVerifyMatchString(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	tests := []struct {
		name     string
		code     string
		contains string
	}{
		{
			name:     "basic",
			code:     `core.verify_match(regex = "TODO")`,
			contains: `regex = "TODO"`,
		},
		{
			name:     "with verify_no_match",
			code:     `core.verify_match(regex = "FIXME", verify_no_match = True)`,
			contains: "verify_no_match = True",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			vm := val.(*core.VerifyMatch)
			str := vm.String()
			if !contains(str, tt.contains) {
				t.Errorf("String() = %q, should contain %q", str, tt.contains)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestVerifyMatchApply(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "verify_match_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files - all files have TODO for the "match found" test
	if err := os.WriteFile(filepath.Join(tmpDir, "has_todo.go"), []byte("// TODO: fix this"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "also_has_todo.go"), []byte("// TODO: another one"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

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
			name:    "match found in all files",
			code:    `core.verify_match(regex = "TODO")`,
			wantErr: false, // all files have TODO
		},
		{
			name:    "no match required - but match exists",
			code:    `core.verify_match(regex = "TODO", verify_no_match = True)`,
			wantErr: true, // fails because TODO is found
		},
		{
			name:    "pattern not found - fails",
			code:    `core.verify_match(regex = "NEVER_EXISTS_PATTERN_XYZ123")`,
			wantErr: true, // fails because pattern not found
		},
		{
			name:    "no match required - pattern not found - passes",
			code:    `core.verify_match(regex = "NEVER_EXISTS_PATTERN_XYZ123", verify_no_match = True)`,
			wantErr: false, // passes because pattern not found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if err != nil {
				t.Fatalf("failed to create verify_match: %v", err)
			}

			vm := val.(*core.VerifyMatch)
			ctx := &transform.Context{WorkDir: tmpDir}

			err = vm.Apply(ctx)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestVerifyMatchWithPaths(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "verify_match_paths_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	if err := os.WriteFile(filepath.Join(tmpDir, "code.go"), []byte("// TODO: fix"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("// TODO: fix"), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	// Only check .txt files - should fail because TODO exists
	val, _ := starlark.Eval(thread, "test.sky",
		`core.verify_match(regex = "TODO", verify_no_match = True, paths = ["*.txt"])`,
		predeclared)

	vm := val.(*core.VerifyMatch)
	ctx := &transform.Context{WorkDir: tmpDir}

	if err := vm.Apply(ctx); err == nil {
		t.Error("expected error because TODO exists in .txt file")
	}

	// Only check .md files (don't exist) - should pass
	val, _ = starlark.Eval(thread, "test.sky",
		`core.verify_match(regex = "TODO", verify_no_match = True, paths = ["*.md"])`,
		predeclared)

	vm = val.(*core.VerifyMatch)
	if err := vm.Apply(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerifyMatchReverse(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	// Without also_on_reversal - reverse should be noop
	val, _ := starlark.Eval(thread, "test.sky", `core.verify_match(regex = "TODO")`, predeclared)
	vm := val.(*core.VerifyMatch)
	reverse := vm.Reverse()

	if reverse.Describe() != "noop" {
		t.Errorf("Reverse().Describe() = %q, want %q", reverse.Describe(), "noop")
	}

	// With also_on_reversal - reverse should be same
	val, _ = starlark.Eval(thread, "test.sky",
		`core.verify_match(regex = "TODO", also_on_reversal = True)`,
		predeclared)
	vm = val.(*core.VerifyMatch)
	reverse = vm.Reverse()

	if reverse.Describe() != "verify_match 'TODO'" {
		t.Errorf("Reverse().Describe() = %q, want %q", reverse.Describe(), "verify_match 'TODO'")
	}
}

func TestVerifyMatchDescribe(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "verify match",
			code:     `core.verify_match(regex = "TODO")`,
			expected: "verify_match 'TODO'",
		},
		{
			name:     "verify no match",
			code:     `core.verify_match(regex = "FIXME", verify_no_match = True)`,
			expected: "verify_no_match 'FIXME'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, _ := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			vm := val.(*core.VerifyMatch)

			if got := vm.Describe(); got != tt.expected {
				t.Errorf("Describe() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestVerifyMatchAccessors(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, _ := starlark.Eval(thread, "test.sky",
		`core.verify_match(regex = "TODO", verify_no_match = True, also_on_reversal = True, failure_message = "Fix this")`,
		predeclared)

	vm := val.(*core.VerifyMatch)

	if vm.RegexString() != "TODO" {
		t.Errorf("RegexString() = %q, want %q", vm.RegexString(), "TODO")
	}

	if !vm.VerifyNoMatch() {
		t.Error("VerifyNoMatch() should be true")
	}

	if !vm.AlsoOnReversal() {
		t.Error("AlsoOnReversal() should be true")
	}

	if vm.FailureMessage() != "Fix this" {
		t.Errorf("FailureMessage() = %q, want %q", vm.FailureMessage(), "Fix this")
	}

	if vm.Regex() == nil {
		t.Error("Regex() should not be nil")
	}

	if vm.Paths() == nil {
		t.Error("Paths() should not be nil")
	}
}

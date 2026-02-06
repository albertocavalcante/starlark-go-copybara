package core_test

import (
	"slices"
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
)

func TestGlobCreation(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		exclude []string
		wantErr bool
	}{
		{
			name:    "simple include",
			include: []string{"**/*.go"},
			exclude: nil,
			wantErr: false,
		},
		{
			name:    "multiple includes",
			include: []string{"*.go", "*.md"},
			exclude: nil,
			wantErr: false,
		},
		{
			name:    "with exclude",
			include: []string{"**"},
			exclude: []string{"*_test.go"},
			wantErr: false,
		},
		{
			name:    "empty include",
			include: []string{},
			exclude: nil,
			wantErr: true,
		},
		{
			name:    "invalid pattern - starts with /",
			include: []string{"/foo/**"},
			exclude: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			glob, err := core.NewGlob(tt.include, tt.exclude)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if glob == nil {
				t.Fatal("expected non-nil glob")
			}
		})
	}
}

func TestGlobMatches(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		exclude []string
		path    string
		want    bool
	}{
		{
			name:    "match all",
			include: []string{"**"},
			path:    "foo/bar/baz.go",
			want:    true,
		},
		{
			name:    "match extension",
			include: []string{"*.go"},
			path:    "main.go",
			want:    true,
		},
		{
			name:    "no match extension",
			include: []string{"*.go"},
			path:    "main.py",
			want:    false,
		},
		{
			name:    "recursive match",
			include: []string{"**/*.go"},
			path:    "foo/bar/main.go",
			want:    true,
		},
		{
			name:    "excluded",
			include: []string{"**"},
			exclude: []string{"*_test.go"},
			path:    "main_test.go",
			want:    false,
		},
		{
			name:    "not excluded",
			include: []string{"**"},
			exclude: []string{"*_test.go"},
			path:    "main.go",
			want:    true,
		},
		{
			name:    "directory pattern",
			include: []string{"foo/**"},
			path:    "foo/bar/baz.txt",
			want:    true,
		},
		{
			name:    "directory pattern no match",
			include: []string{"foo/**"},
			path:    "bar/baz.txt",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			glob, err := core.NewGlob(tt.include, tt.exclude)
			if err != nil {
				t.Fatalf("failed to create glob: %v", err)
			}

			got := glob.Matches(tt.path)
			if got != tt.want {
				t.Errorf("Matches(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGlobString(t *testing.T) {
	tests := []struct {
		name     string
		include  []string
		exclude  []string
		expected string
	}{
		{
			name:     "simple",
			include:  []string{"**"},
			expected: `glob(include = ["**"])`,
		},
		{
			name:     "with exclude",
			include:  []string{"**"},
			exclude:  []string{"*_test.go"},
			expected: `glob(include = ["**"], exclude = ["*_test.go"])`,
		},
		{
			name:     "multiple patterns",
			include:  []string{"*.go", "*.md"},
			expected: `glob(include = ["*.go", "*.md"])`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			glob, err := core.NewGlob(tt.include, tt.exclude)
			if err != nil {
				t.Fatalf("failed to create glob: %v", err)
			}

			if got := glob.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGlobStarlarkValue(t *testing.T) {
	glob, err := core.NewGlob([]string{"**"}, nil)
	if err != nil {
		t.Fatalf("failed to create glob: %v", err)
	}

	// Test Type
	if got := glob.Type(); got != "glob" {
		t.Errorf("Type() = %q, want %q", got, "glob")
	}

	// Test Truth
	if got := glob.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	// Test Hash (should error)
	if _, err := glob.Hash(); err == nil {
		t.Error("Hash() should return error")
	}

	// Test Freeze (should not panic)
	glob.Freeze()
}

func TestGlobUnion(t *testing.T) {
	g1, _ := core.NewGlob([]string{"*.go"}, nil)
	g2, _ := core.NewGlob([]string{"*.md"}, nil)

	union := core.Union(g1, g2)

	// Should match both patterns
	if !union.Matches("main.go") {
		t.Error("union should match main.go")
	}
	if !union.Matches("README.md") {
		t.Error("union should match README.md")
	}
	if union.Matches("main.py") {
		t.Error("union should not match main.py")
	}
}

func TestGlobDifference(t *testing.T) {
	g1, _ := core.NewGlob([]string{"**"}, nil)
	g2, _ := core.NewGlob([]string{"*_test.go"}, nil)

	diff := core.Difference(g1, g2)

	// Should match non-test files
	if !diff.Matches("main.go") {
		t.Error("difference should match main.go")
	}
	if diff.Matches("main_test.go") {
		t.Error("difference should not match main_test.go")
	}
}

func TestGlobRoots(t *testing.T) {
	tests := []struct {
		name    string
		include []string
		want    []string
	}{
		{
			name:    "all files",
			include: []string{"**"},
			want:    []string{""},
		},
		{
			name:    "single directory",
			include: []string{"foo/**"},
			want:    []string{"foo"},
		},
		{
			name:    "multiple directories",
			include: []string{"foo/**", "bar/**"},
			want:    []string{"bar", "foo"},
		},
		{
			name:    "nested",
			include: []string{"foo/bar/**"},
			want:    []string{"foo/bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			glob, _ := core.NewGlob(tt.include, nil)
			got := glob.Roots()
			slices.Sort(got)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Roots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlobAllFiles(t *testing.T) {
	glob := core.AllFiles()

	if !glob.IsAllFiles() {
		t.Error("AllFiles() should return glob that IsAllFiles()")
	}

	if !glob.Matches("any/path/file.txt") {
		t.Error("AllFiles() should match any path")
	}
}

func TestGlobFromStarlark(t *testing.T) {
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
			name:    "core.glob basic",
			code:    `core.glob(include = ["**/*.go"])`,
			wantErr: false,
		},
		{
			name:    "core.glob with exclude",
			code:    `core.glob(include = ["**"], exclude = ["*_test.go"])`,
			wantErr: false,
		},
		{
			name:    "global glob function",
			code:    `glob(["**/*.go"])`,
			wantErr: false,
		},
		{
			name:    "glob with exclude",
			code:    `glob(["**"], exclude = ["vendor/**"])`,
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

			glob, ok := val.(*core.Glob)
			if !ok {
				t.Fatalf("expected *Glob, got %T", val)
			}

			if glob.Type() != "glob" {
				t.Errorf("Type() = %q, want %q", glob.Type(), "glob")
			}
		})
	}
}

func TestGlobBinaryOps(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"glob": core.Globals()["glob"],
	}

	tests := []struct {
		name     string
		code     string
		testPath string
		want     bool
	}{
		{
			name:     "union with +",
			code:     `glob(["*.go"]) + glob(["*.md"])`,
			testPath: "README.md",
			want:     true,
		},
		{
			name:     "difference with -",
			code:     `glob(["**"]) - glob(["*_test.go"])`,
			testPath: "main_test.go",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			glob, ok := val.(*core.Glob)
			if !ok {
				t.Fatalf("expected *Glob, got %T", val)
			}

			if got := glob.Matches(tt.testPath); got != tt.want {
				t.Errorf("Matches(%q) = %v, want %v", tt.testPath, got, tt.want)
			}
		})
	}
}

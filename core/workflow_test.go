package core_test

import (
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
)

func TestWorkflowCreation(t *testing.T) {
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
			name:    "minimal workflow",
			code:    `core.workflow(name = "default")`,
			wantErr: false,
		},
		{
			name:    "workflow with mode",
			code:    `core.workflow(name = "default", mode = "ITERATIVE")`,
			wantErr: false,
		},
		{
			name:    "workflow with CHANGE_REQUEST mode",
			code:    `core.workflow(name = "default", mode = "CHANGE_REQUEST")`,
			wantErr: false,
		},
		{
			name:    "workflow with origin_files glob",
			code:    `core.workflow(name = "default", origin_files = core.glob(include = ["**/*.go"]))`,
			wantErr: false,
		},
		{
			name:    "workflow with origin_files list",
			code:    `core.workflow(name = "default", origin_files = ["**/*.go"])`,
			wantErr: false,
		},
		{
			name:    "workflow with destination_files",
			code:    `core.workflow(name = "default", destination_files = ["**"])`,
			wantErr: false,
		},
		{
			name:    "workflow with reversible_check",
			code:    `core.workflow(name = "default", reversible_check = True)`,
			wantErr: false,
		},
		{
			name:    "workflow with transformations",
			code:    `core.workflow(name = "default", transformations = [core.move("src", "lib")])`,
			wantErr: false,
		},
		{
			name:    "invalid mode",
			code:    `core.workflow(name = "default", mode = "INVALID")`,
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

			wf, ok := val.(*core.Workflow)
			if !ok {
				t.Fatalf("expected *Workflow, got %T", val)
			}

			if wf.Name() != "default" {
				t.Errorf("Name() = %q, want %q", wf.Name(), "default")
			}
		})
	}
}

func TestWorkflowModes(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	tests := []struct {
		mode     string
		expected core.WorkflowMode
	}{
		{"SQUASH", core.ModeSquash},
		{"ITERATIVE", core.ModeIterative},
		{"CHANGE_REQUEST", core.ModeChangeRequest},
		{"CHANGE_REQUEST_FROM_SOT", core.ModeChangeRequestFromSOT},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			code := `core.workflow(name = "test", mode = "` + tt.mode + `")`
			val, err := starlark.Eval(thread, "test.sky", code, predeclared)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wf := val.(*core.Workflow)
			if wf.Mode() != tt.expected {
				t.Errorf("Mode() = %v, want %v", wf.Mode(), tt.expected)
			}
		})
	}
}

func TestWorkflowModeString(t *testing.T) {
	tests := []struct {
		mode     core.WorkflowMode
		expected string
	}{
		{core.ModeSquash, "SQUASH"},
		{core.ModeIterative, "ITERATIVE"},
		{core.ModeChangeRequest, "CHANGE_REQUEST"},
		{core.ModeChangeRequestFromSOT, "CHANGE_REQUEST_FROM_SOT"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestWorkflowReversibleCheckDefault(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"SQUASH defaults to false", "SQUASH", false},
		{"ITERATIVE defaults to false", "ITERATIVE", false},
		{"CHANGE_REQUEST defaults to true", "CHANGE_REQUEST", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := `core.workflow(name = "test", mode = "` + tt.mode + `")`
			val, err := starlark.Eval(thread, "test.sky", code, predeclared)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wf := val.(*core.Workflow)
			if wf.ReversibleCheck() != tt.expected {
				t.Errorf("ReversibleCheck() = %v, want %v", wf.ReversibleCheck(), tt.expected)
			}
		})
	}
}

func TestWorkflowOriginFiles(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
		"glob": core.Globals()["glob"],
	}

	tests := []struct {
		name         string
		code         string
		wantAllFiles bool
	}{
		{
			name:         "default is all files",
			code:         `core.workflow(name = "test")`,
			wantAllFiles: true,
		},
		{
			name:         "with glob",
			code:         `core.workflow(name = "test", origin_files = glob(["*.go"]))`,
			wantAllFiles: false,
		},
		{
			name:         "with list",
			code:         `core.workflow(name = "test", origin_files = ["*.go"])`,
			wantAllFiles: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wf := val.(*core.Workflow)
			originFiles := wf.OriginFiles()

			if originFiles == nil {
				t.Fatal("OriginFiles() should not be nil")
			}

			if originFiles.IsAllFiles() != tt.wantAllFiles {
				t.Errorf("OriginFiles().IsAllFiles() = %v, want %v",
					originFiles.IsAllFiles(), tt.wantAllFiles)
			}
		})
	}
}

func TestWorkflowTransformations(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	code := `core.workflow(
		name = "test",
		transformations = [
			core.move("src", "lib"),
			core.replace("foo", "bar"),
		]
	)`

	val, err := starlark.Eval(thread, "test.sky", code, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf := val.(*core.Workflow)
	transforms := wf.Transformations()

	if len(transforms) != 2 {
		t.Fatalf("len(Transformations()) = %d, want 2", len(transforms))
	}

	if transforms[0].Type() != "move" {
		t.Errorf("transforms[0].Type() = %q, want %q", transforms[0].Type(), "move")
	}

	if transforms[1].Type() != "replace" {
		t.Errorf("transforms[1].Type() = %q, want %q", transforms[1].Type(), "replace")
	}
}

func TestWorkflowStarlarkValue(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `core.workflow(name = "test")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf := val.(*core.Workflow)

	// Test Type
	if got := wf.Type(); got != "workflow" {
		t.Errorf("Type() = %q, want %q", got, "workflow")
	}

	// Test String
	expected := `workflow("test")`
	if got := wf.String(); got != expected {
		t.Errorf("String() = %q, want %q", got, expected)
	}

	// Test Truth
	if got := wf.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	// Test Hash (should error)
	if _, err := wf.Hash(); err == nil {
		t.Error("Hash() should return error")
	}

	// Test Freeze (should not panic)
	wf.Freeze()
}

func TestWorkflowAccessors(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"core": core.Module,
	}

	code := `core.workflow(
		name = "myworkflow",
		mode = "ITERATIVE",
		reversible_check = True
	)`

	val, err := starlark.Eval(thread, "test.sky", code, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wf := val.(*core.Workflow)

	if wf.Name() != "myworkflow" {
		t.Errorf("Name() = %q, want %q", wf.Name(), "myworkflow")
	}

	if wf.Mode() != core.ModeIterative {
		t.Errorf("Mode() = %v, want %v", wf.Mode(), core.ModeIterative)
	}

	if !wf.ReversibleCheck() {
		t.Error("ReversibleCheck() should be true")
	}

	if wf.OriginFiles() == nil {
		t.Error("OriginFiles() should not be nil")
	}

	if wf.DestinationFiles() == nil {
		t.Error("DestinationFiles() should not be nil")
	}
}

func TestParseWorkflowMode(t *testing.T) {
	tests := []struct {
		input    string
		expected core.WorkflowMode
		wantErr  bool
	}{
		{"SQUASH", core.ModeSquash, false},
		{"squash", core.ModeSquash, false},
		{"Squash", core.ModeSquash, false},
		{"ITERATIVE", core.ModeIterative, false},
		{"CHANGE_REQUEST", core.ModeChangeRequest, false},
		{"CHANGE_REQUEST_FROM_SOT", core.ModeChangeRequestFromSOT, false},
		{"INVALID", core.ModeSquash, true},
		{"", core.ModeSquash, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mode, err := core.ParseWorkflowMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mode != tt.expected {
				t.Errorf("ParseWorkflowMode(%q) = %v, want %v", tt.input, mode, tt.expected)
			}
		})
	}
}

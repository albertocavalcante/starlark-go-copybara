package copybara_test

import (
	"testing"

	"github.com/albertocavalcante/starlark-go-copybara/copybara"
)

func TestNew(t *testing.T) {
	interp := copybara.New()
	if interp == nil {
		t.Fatal("expected non-nil interpreter")
	}
}

func TestEval_EmptyConfig(t *testing.T) {
	interp := copybara.New()

	result, err := interp.Eval("copy.bara.sky", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	workflows := result.Workflows()
	if len(workflows) != 0 {
		t.Errorf("expected 0 workflows, got %d", len(workflows))
	}
}

func TestModulesRegistered(t *testing.T) {
	interp := copybara.New()

	// Test that all modules are registered by evaluating code that uses them
	tests := []struct {
		name   string
		config string
	}{
		{
			name:   "core module",
			config: `_ = core.workflow(name = "test")`,
		},
		{
			name:   "git module",
			config: `_ = git.origin(url = "https://github.com/test/repo")`,
		},
		{
			name:   "metadata module",
			config: `_ = metadata.squash_notes()`,
		},
		{
			name:   "authoring module",
			config: `_ = authoring.overwrite(default = "Test <test@example.com>")`,
		},
		{
			name:   "folder module",
			config: `_ = folder.origin()`,
		},
		{
			name:   "glob global function",
			config: `_ = glob(["**/*.go"])`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := interp.Eval("copy.bara.sky", tt.config)
			if err != nil {
				t.Errorf("module not registered properly: %v", err)
			}
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithDryRun", func(t *testing.T) {
		interp := copybara.New(copybara.WithDryRun(true))
		if !interp.DryRun() {
			t.Error("expected DryRun to be true")
		}

		interp2 := copybara.New(copybara.WithDryRun(false))
		if interp2.DryRun() {
			t.Error("expected DryRun to be false")
		}
	})

	t.Run("WithWorkdir", func(t *testing.T) {
		interp := copybara.New(copybara.WithWorkdir("/tmp/test"))
		if interp.WorkDir() != "/tmp/test" {
			t.Errorf("expected WorkDir to be '/tmp/test', got %q", interp.WorkDir())
		}
	})

	t.Run("multiple options", func(t *testing.T) {
		interp := copybara.New(
			copybara.WithDryRun(true),
			copybara.WithWorkdir("/custom/path"),
		)
		if !interp.DryRun() {
			t.Error("expected DryRun to be true")
		}
		if interp.WorkDir() != "/custom/path" {
			t.Errorf("expected WorkDir to be '/custom/path', got %q", interp.WorkDir())
		}
	})
}

func TestEval_ExtractsWorkflows(t *testing.T) {
	interp := copybara.New()

	config := `
my_workflow = core.workflow(
    name = "default",
    origin = folder.origin(),
    destination = folder.destination(),
    authoring = authoring.overwrite(default = "Test <test@example.com>"),
)
`

	result, err := interp.Eval("copy.bara.sky", config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	workflows := result.Workflows()
	if len(workflows) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(workflows))
	}

	if workflows[0].Name() != "default" {
		t.Errorf("expected workflow name 'default', got %q", workflows[0].Name())
	}
}

func TestEval_MultipleWorkflows(t *testing.T) {
	interp := copybara.New()

	config := `
wf1 = core.workflow(
    name = "workflow1",
    origin = folder.origin(),
    destination = folder.destination(),
    authoring = authoring.overwrite(default = "Test <test@example.com>"),
)

wf2 = core.workflow(
    name = "workflow2",
    origin = folder.origin(),
    destination = folder.destination(),
    authoring = authoring.overwrite(default = "Test <test@example.com>"),
)
`

	result, err := interp.Eval("copy.bara.sky", config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	workflows := result.Workflows()
	if len(workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(workflows))
	}

	// Check that both workflows are present (order may vary)
	names := make(map[string]bool)
	for _, wf := range workflows {
		names[wf.Name()] = true
	}

	if !names["workflow1"] {
		t.Error("expected to find workflow1")
	}
	if !names["workflow2"] {
		t.Error("expected to find workflow2")
	}
}

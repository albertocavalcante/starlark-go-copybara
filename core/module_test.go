package core_test

import (
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/core"
)

func TestModule(t *testing.T) {
	if core.Module == nil {
		t.Fatal("expected non-nil module")
	}

	if core.Module.Name != "core" {
		t.Errorf("expected module name 'core', got %q", core.Module.Name)
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

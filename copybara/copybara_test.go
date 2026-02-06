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

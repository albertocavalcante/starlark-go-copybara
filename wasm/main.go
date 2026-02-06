//go:build js && wasm

// Package main provides the WASM entry point for starlark-go-copybara.
package main

import (
	"syscall/js"

	"github.com/albertocavalcante/starlark-go-copybara/copybara"
)

func main() {
	// Register JavaScript API
	js.Global().Set("Copybara", map[string]any{
		"eval":    js.FuncOf(evalConfig),
		"version": js.ValueOf("0.1.0"),
	})

	// Keep the program running
	select {}
}

// evalConfig evaluates a Copybara configuration.
func evalConfig(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return map[string]any{
			"error": "usage: Copybara.eval(filename, source)",
		}
	}

	filename := args[0].String()
	source := args[1].String()

	interp := copybara.New()
	result, err := interp.Eval(filename, source)
	if err != nil {
		return map[string]any{
			"error": err.Error(),
		}
	}

	workflows := result.Workflows()
	workflowNames := make([]any, len(workflows))
	for i, wf := range workflows {
		workflowNames[i] = wf.Name()
	}

	return map[string]any{
		"workflows": workflowNames,
	}
}

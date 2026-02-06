package core

import (
	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// Transformation is the interface for all code transformations.
// All transformations must also implement starlark.Value.
// This embeds transform.Transformation and adds the starlark.Value requirement.
type Transformation interface {
	starlark.Value
	transform.Transformation
}

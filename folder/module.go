// Package folder provides the folder.* Starlark module for local folder operations.
//
// The folder module is primarily used for testing workflows locally:
//   - folder.origin() - Local folder origin
//   - folder.destination() - Local folder destination
//
// Reference: https://github.com/google/copybara/tree/master/java/com/google/copybara/folder
package folder

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the folder.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "folder",
	Members: starlark.StringDict{
		"origin":      starlark.NewBuiltin("folder.origin", originFn),
		"destination": starlark.NewBuiltin("folder.destination", destinationFn),
	},
}

// Origin represents a local folder origin.
type Origin struct {
	path string
}

// String implements starlark.Value.
func (o *Origin) String() string {
	return fmt.Sprintf("folder.origin(path = %q)", o.path)
}

// Type implements starlark.Value.
func (o *Origin) Type() string {
	return "folder.origin"
}

// Freeze implements starlark.Value.
func (o *Origin) Freeze() {}

// Truth implements starlark.Value.
func (o *Origin) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (o *Origin) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: folder.origin")
}

// Destination represents a local folder destination.
type Destination struct {
	path string
}

// String implements starlark.Value.
func (d *Destination) String() string {
	return fmt.Sprintf("folder.destination(path = %q)", d.path)
}

// Type implements starlark.Value.
func (d *Destination) Type() string {
	return "folder.destination"
}

// Freeze implements starlark.Value.
func (d *Destination) Freeze() {}

// Truth implements starlark.Value.
func (d *Destination) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (d *Destination) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: folder.destination")
}

// originFn implements folder.origin().
func originFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"path?", &path,
	); err != nil {
		return nil, err
	}

	return &Origin{path: path}, nil
}

// destinationFn implements folder.destination().
func destinationFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"path?", &path,
	); err != nil {
		return nil, err
	}

	return &Destination{path: path}, nil
}

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
	"os"

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
	path                       string
	materializeOutsideSymlinks bool
}

// String implements starlark.Value.
func (o *Origin) String() string {
	if o.path == "" {
		return "folder.origin()"
	}
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

// Path returns the origin path.
func (o *Origin) Path() string {
	if o.path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return o.path
}

// MaterializeOutsideSymlinks returns whether symlinks outside the origin should be materialized.
func (o *Origin) MaterializeOutsideSymlinks() bool {
	return o.materializeOutsideSymlinks
}

// Attr implements starlark.HasAttrs for accessing origin properties.
func (o *Origin) Attr(name string) (starlark.Value, error) {
	switch name {
	case "path":
		return starlark.String(o.Path()), nil
	case "materialize_outside_symlinks":
		return starlark.Bool(o.materializeOutsideSymlinks), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (o *Origin) AttrNames() []string {
	return []string{"path", "materialize_outside_symlinks"}
}

// Destination represents a local folder destination.
type Destination struct {
	path string
}

// String implements starlark.Value.
func (d *Destination) String() string {
	if d.path == "" {
		return "folder.destination()"
	}
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

// Path returns the destination path.
func (d *Destination) Path() string {
	if d.path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "."
		}
		return wd
	}
	return d.path
}

// Attr implements starlark.HasAttrs for accessing destination properties.
func (d *Destination) Attr(name string) (starlark.Value, error) {
	switch name {
	case "path":
		return starlark.String(d.Path()), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (d *Destination) AttrNames() []string {
	return []string{"path"}
}

// originFn implements folder.origin().
//
// Parameters:
//   - path (optional): The local folder path. Defaults to the current working directory.
//   - materialize_outside_symlinks (optional): If true, symlinks pointing outside the
//     origin directory will be materialized (their targets will be copied). Defaults to false.
func originFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		path                       string
		materializeOutsideSymlinks bool
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"path?", &path,
		"materialize_outside_symlinks?", &materializeOutsideSymlinks,
	); err != nil {
		return nil, err
	}

	return &Origin{
		path:                       path,
		materializeOutsideSymlinks: materializeOutsideSymlinks,
	}, nil
}

// destinationFn implements folder.destination().
//
// Parameters:
//   - path (optional): The local folder path. Defaults to the current working directory.
func destinationFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var path string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"path?", &path,
	); err != nil {
		return nil, err
	}

	return &Destination{path: path}, nil
}

// Impl returns an implementation of the Origin that can be used for file operations.
func (o *Origin) Impl() *OriginImpl {
	return NewOriginImpl(o)
}

// Impl returns an implementation of the Destination that can be used for file operations.
func (d *Destination) Impl() *DestinationImpl {
	return NewDestinationImpl(d)
}

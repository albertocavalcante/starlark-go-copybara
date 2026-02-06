// Package authoring provides the authoring.* Starlark module for author handling.
//
// The authoring module provides functions for handling commit authoring:
//   - authoring.pass_thru() - Pass through authoring unchanged
//   - authoring.overwrite() - Overwrite with a fixed author
//   - authoring.allowed() - Allow only listed authors
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/authoring/Authoring.java
package authoring

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the authoring.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "authoring",
	Members: starlark.StringDict{
		"pass_thru": starlark.NewBuiltin("authoring.pass_thru", passThruFn),
		"overwrite": starlark.NewBuiltin("authoring.overwrite", overwriteFn),
		"allowed":   starlark.NewBuiltin("authoring.allowed", allowedFn),
	},
}

// PassThru represents pass-through authoring.
type PassThru struct {
	defaultAuthor string
}

func (p *PassThru) String() string        { return "authoring.pass_thru()" }
func (p *PassThru) Type() string          { return "pass_thru" }
func (p *PassThru) Freeze()               {}
func (p *PassThru) Truth() starlark.Bool  { return starlark.True }
func (p *PassThru) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: pass_thru") }

// Overwrite represents overwrite authoring.
type Overwrite struct {
	author string
}

func (o *Overwrite) String() string        { return fmt.Sprintf("authoring.overwrite(%q)", o.author) }
func (o *Overwrite) Type() string          { return "overwrite" }
func (o *Overwrite) Freeze()               {}
func (o *Overwrite) Truth() starlark.Bool  { return starlark.True }
func (o *Overwrite) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: overwrite") }

// Allowed represents allowed authoring with a whitelist.
type Allowed struct {
	defaultAuthor string
	allowlist     []string
}

func (a *Allowed) String() string        { return "authoring.allowed()" }
func (a *Allowed) Type() string          { return "allowed" }
func (a *Allowed) Freeze()               {}
func (a *Allowed) Truth() starlark.Bool  { return starlark.True }
func (a *Allowed) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: allowed") }

// passThruFn implements authoring.pass_thru().
func passThruFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var defaultAuthor string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"default?", &defaultAuthor,
	); err != nil {
		return nil, err
	}

	return &PassThru{defaultAuthor: defaultAuthor}, nil
}

// overwriteFn implements authoring.overwrite().
func overwriteFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var author string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"default", &author,
	); err != nil {
		return nil, err
	}

	return &Overwrite{author: author}, nil
}

// allowedFn implements authoring.allowed().
func allowedFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var defaultAuthor string
	var allowlist *starlark.List

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"default", &defaultAuthor,
		"allowlist?", &allowlist,
	); err != nil {
		return nil, err
	}

	var authors []string
	if allowlist != nil {
		for i := range allowlist.Len() {
			if s, ok := allowlist.Index(i).(starlark.String); ok {
				authors = append(authors, string(s))
			}
		}
	}

	return &Allowed{
		defaultAuthor: defaultAuthor,
		allowlist:     authors,
	}, nil
}

// Package git provides the git.* Starlark module for Copybara.
//
// The git module provides origins and destinations for Git repositories:
//   - git.origin() - Git repository origin
//   - git.destination() - Git repository destination
//   - git.github_origin() - GitHub origin
//   - git.github_pr_destination() - GitHub PR destination
//
// Reference: https://github.com/google/copybara/tree/master/java/com/google/copybara/git
package git

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the git.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "git",
	Members: starlark.StringDict{
		"origin":      starlark.NewBuiltin("git.origin", originFn),
		"destination": starlark.NewBuiltin("git.destination", destinationFn),
	},
}

// originFn implements git.origin().
//
// Reference: GitOrigin.java
func originFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url string
		ref string
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"url", &url,
		"ref?", &ref,
	); err != nil {
		return nil, err
	}

	return &Origin{
		url: url,
		ref: ref,
	}, nil
}

// destinationFn implements git.destination().
//
// Reference: GitDestination.java
func destinationFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url    string
		push   string
		fetch  string
		branch string
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"url", &url,
		"push?", &push,
		"fetch?", &fetch,
		"branch?", &branch,
	); err != nil {
		return nil, err
	}

	// Default branch to "main" if not specified
	if push == "" && branch != "" {
		push = branch
	}
	if fetch == "" && branch != "" {
		fetch = branch
	}

	return &Destination{
		url:   url,
		push:  push,
		fetch: fetch,
	}, nil
}

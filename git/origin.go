// Package git provides the git.* Starlark module for Copybara.
package git

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// SubmoduleStrategy defines how submodules should be handled.
type SubmoduleStrategy string

const (
	// SubmoduleNo disables submodule downloading.
	SubmoduleNo SubmoduleStrategy = "NO"
	// SubmoduleYes downloads first level submodules.
	SubmoduleYes SubmoduleStrategy = "YES"
	// SubmoduleRecursive downloads all submodules recursively.
	SubmoduleRecursive SubmoduleStrategy = "RECURSIVE"
)

// ParseSubmoduleStrategy parses a string into a SubmoduleStrategy.
func ParseSubmoduleStrategy(s string) (SubmoduleStrategy, error) {
	switch strings.ToUpper(s) {
	case "NO", "":
		return SubmoduleNo, nil
	case "YES":
		return SubmoduleYes, nil
	case "RECURSIVE":
		return SubmoduleRecursive, nil
	default:
		return "", fmt.Errorf("invalid submodule strategy: %q (expected NO, YES, or RECURSIVE)", s)
	}
}

// Origin represents a Git repository origin.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitOrigin.java
type Origin struct {
	url                      string
	ref                      string
	submodules               SubmoduleStrategy
	includeBranchCommitLogs  bool
	firstParent              bool
	partialFetch             bool
	primaryBranchMigration   bool
}

// String implements starlark.Value.
func (o *Origin) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("url = %q", o.url))
	if o.ref != "" {
		parts = append(parts, fmt.Sprintf("ref = %q", o.ref))
	}
	if o.submodules != "" && o.submodules != SubmoduleNo {
		parts = append(parts, fmt.Sprintf("submodules = %q", o.submodules))
	}
	if o.includeBranchCommitLogs {
		parts = append(parts, "include_branch_commit_logs = True")
	}
	if !o.firstParent {
		parts = append(parts, "first_parent = False")
	}
	if o.partialFetch {
		parts = append(parts, "partial_fetch = True")
	}
	return fmt.Sprintf("git.origin(%s)", strings.Join(parts, ", "))
}

// Type implements starlark.Value.
func (o *Origin) Type() string {
	return "git.origin"
}

// Freeze implements starlark.Value.
func (o *Origin) Freeze() {}

// Truth implements starlark.Value.
func (o *Origin) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (o *Origin) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.origin")
}

// URL returns the repository URL.
func (o *Origin) URL() string {
	return o.url
}

// Ref returns the Git reference.
func (o *Origin) Ref() string {
	return o.ref
}

// Submodules returns the submodule strategy.
func (o *Origin) Submodules() SubmoduleStrategy {
	return o.submodules
}

// IncludeBranchCommitLogs returns whether to include branch commit logs.
func (o *Origin) IncludeBranchCommitLogs() bool {
	return o.includeBranchCommitLogs
}

// FirstParent returns whether to use first parent only.
func (o *Origin) FirstParent() bool {
	return o.firstParent
}

// PartialFetch returns whether partial fetch is enabled.
func (o *Origin) PartialFetch() bool {
	return o.partialFetch
}

// PrimaryBranchMigration returns whether primary branch migration mode is enabled.
func (o *Origin) PrimaryBranchMigration() bool {
	return o.primaryBranchMigration
}

// Attr implements starlark.HasAttrs.
func (o *Origin) Attr(name string) (starlark.Value, error) {
	switch name {
	case "url":
		return starlark.String(o.url), nil
	case "ref":
		return starlark.String(o.ref), nil
	case "submodules":
		return starlark.String(o.submodules), nil
	case "include_branch_commit_logs":
		return starlark.Bool(o.includeBranchCommitLogs), nil
	case "first_parent":
		return starlark.Bool(o.firstParent), nil
	case "partial_fetch":
		return starlark.Bool(o.partialFetch), nil
	case "primary_branch_migration":
		return starlark.Bool(o.primaryBranchMigration), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (o *Origin) AttrNames() []string {
	return []string{
		"url",
		"ref",
		"submodules",
		"include_branch_commit_logs",
		"first_parent",
		"partial_fetch",
		"primary_branch_migration",
	}
}

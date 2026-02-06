package git

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// Destination represents a Git repository destination.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitDestination.java
type Destination struct {
	url                    string
	push                   string
	fetch                  string
	tagName                string
	tagMsg                 string
	skipPush               bool
	primaryBranchMigration bool
	integrates             []*IntegrateChanges
}

// String implements starlark.Value.
func (d *Destination) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("url = %q", d.url))
	if d.push != "" {
		parts = append(parts, fmt.Sprintf("push = %q", d.push))
	}
	if d.fetch != "" {
		parts = append(parts, fmt.Sprintf("fetch = %q", d.fetch))
	}
	if d.tagName != "" {
		parts = append(parts, fmt.Sprintf("tag_name = %q", d.tagName))
	}
	if d.tagMsg != "" {
		parts = append(parts, fmt.Sprintf("tag_msg = %q", d.tagMsg))
	}
	if d.skipPush {
		parts = append(parts, "skip_push = True")
	}
	return fmt.Sprintf("git.destination(%s)", strings.Join(parts, ", "))
}

// Type implements starlark.Value.
func (d *Destination) Type() string {
	return "git.destination"
}

// Freeze implements starlark.Value.
func (d *Destination) Freeze() {}

// Truth implements starlark.Value.
func (d *Destination) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (d *Destination) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.destination")
}

// URL returns the repository URL.
func (d *Destination) URL() string {
	return d.url
}

// Push returns the push branch.
func (d *Destination) Push() string {
	return d.push
}

// Fetch returns the fetch branch.
func (d *Destination) Fetch() string {
	return d.fetch
}

// TagName returns the tag name template.
func (d *Destination) TagName() string {
	return d.tagName
}

// TagMsg returns the tag message template.
func (d *Destination) TagMsg() string {
	return d.tagMsg
}

// SkipPush returns whether push should be skipped (dry-run mode).
func (d *Destination) SkipPush() bool {
	return d.skipPush
}

// PrimaryBranchMigration returns whether primary branch migration mode is enabled.
func (d *Destination) PrimaryBranchMigration() bool {
	return d.primaryBranchMigration
}

// Integrates returns the integrate changes configuration.
func (d *Destination) Integrates() []*IntegrateChanges {
	return d.integrates
}

// Attr implements starlark.HasAttrs.
func (d *Destination) Attr(name string) (starlark.Value, error) {
	switch name {
	case "url":
		return starlark.String(d.url), nil
	case "push":
		return starlark.String(d.push), nil
	case "fetch":
		return starlark.String(d.fetch), nil
	case "tag_name":
		return starlark.String(d.tagName), nil
	case "tag_msg":
		return starlark.String(d.tagMsg), nil
	case "skip_push":
		return starlark.Bool(d.skipPush), nil
	case "primary_branch_migration":
		return starlark.Bool(d.primaryBranchMigration), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (d *Destination) AttrNames() []string {
	return []string{
		"url",
		"push",
		"fetch",
		"tag_name",
		"tag_msg",
		"skip_push",
		"primary_branch_migration",
	}
}

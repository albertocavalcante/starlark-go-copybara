package types

import (
	"fmt"
	"regexp"

	"go.starlark.net/starlark"
)

// Author represents a commit author.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Author.java
type Author struct {
	name  string
	email string
}

// authorPattern matches "Name <email>" format.
var authorPattern = regexp.MustCompile(`^(.+)\s+<(.+)>$`)

// ParseAuthor parses an author string in "Name <email>" format.
func ParseAuthor(s string) (*Author, error) {
	matches := authorPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid author format: %q (expected 'Name <email>')", s)
	}
	return &Author{
		name:  matches[1],
		email: matches[2],
	}, nil
}

// NewAuthor creates a new Author.
func NewAuthor(name, email string) *Author {
	return &Author{name: name, email: email}
}

// String implements starlark.Value.
func (a *Author) String() string {
	return fmt.Sprintf("%s <%s>", a.name, a.email)
}

// Type implements starlark.Value.
func (a *Author) Type() string {
	return "author"
}

// Freeze implements starlark.Value.
func (a *Author) Freeze() {}

// Truth implements starlark.Value.
func (a *Author) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (a *Author) Hash() (uint32, error) {
	return starlark.String(a.String()).Hash()
}

// Name returns the author's name.
func (a *Author) Name() string {
	return a.name
}

// Email returns the author's email.
func (a *Author) Email() string {
	return a.email
}

// Attr implements starlark.HasAttrs.
func (a *Author) Attr(name string) (starlark.Value, error) {
	switch name {
	case "name":
		return starlark.String(a.name), nil
	case "email":
		return starlark.String(a.email), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (a *Author) AttrNames() []string {
	return []string{"name", "email"}
}

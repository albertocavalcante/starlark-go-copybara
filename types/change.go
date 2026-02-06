package types

import (
	"fmt"
	"time"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/authoring"
)

// Ensure Change implements required interfaces.
var (
	_ starlark.Value    = (*Change)(nil)
	_ starlark.HasAttrs = (*Change)(nil)
)

// Change represents a VCS change/commit.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/Change.java
type Change struct {
	ref      string
	author   *authoring.Author
	message  *ChangeMessage
	dateTime time.Time
	files    []string
}

// NewChange creates a new Change.
func NewChange(ref string, author *authoring.Author, message string, dateTime time.Time, files []string) *Change {
	return &Change{
		ref:      ref,
		author:   author,
		message:  NewChangeMessage(message),
		dateTime: dateTime,
		files:    files,
	}
}

// String implements starlark.Value.
func (c *Change) String() string {
	return fmt.Sprintf("change<%s>", c.ref)
}

// Type implements starlark.Value.
func (c *Change) Type() string {
	return "change"
}

// Freeze implements starlark.Value.
func (c *Change) Freeze() {}

// Truth implements starlark.Value.
func (c *Change) Truth() starlark.Bool {
	return starlark.Bool(c.ref != "")
}

// Hash implements starlark.Value.
func (c *Change) Hash() (uint32, error) {
	return starlark.String(c.ref).Hash()
}

// Ref returns the change reference (commit hash, revision number, etc.).
func (c *Change) Ref() string {
	return c.ref
}

// Author returns the change author.
func (c *Change) Author() *authoring.Author {
	return c.author
}

// Message returns the change message.
func (c *Change) Message() *ChangeMessage {
	return c.message
}

// DateTime returns the change timestamp.
func (c *Change) DateTime() time.Time {
	return c.dateTime
}

// Files returns the list of files changed.
func (c *Change) Files() []string {
	return c.files
}

// Attr implements starlark.HasAttrs.
func (c *Change) Attr(name string) (starlark.Value, error) {
	switch name {
	case "ref":
		return starlark.String(c.ref), nil
	case "author":
		return c.author, nil
	case "message":
		return c.message, nil
	case "date_time":
		return starlark.String(c.dateTime.Format(time.RFC3339)), nil
	case "files":
		return StringSliceToList(c.files), nil
	default:
		return nil, nil
	}
}

// StringSliceToList converts a slice of strings to a starlark.List.
func StringSliceToList(ss []string) *starlark.List {
	elems := make([]starlark.Value, len(ss))
	for i, s := range ss {
		elems[i] = starlark.String(s)
	}
	return starlark.NewList(elems)
}

// AttrNames implements starlark.HasAttrs.
func (c *Change) AttrNames() []string {
	return []string{"author", "date_time", "files", "message", "ref"}
}

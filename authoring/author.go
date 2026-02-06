// Package authoring provides Author type and parsing for commit authoring.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/authoring/Author.java
package authoring

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.starlark.net/starlark"
)

// authorPattern matches the standard "Name <email>" format.
// The name can contain any characters except <, and email can contain any characters except >.
var authorPattern = regexp.MustCompile(`^(?P<name>[^<]+)<(?P<email>[^>]*)>$`)

// doubleQuotedPattern matches strings wrapped in double quotes.
var doubleQuotedPattern = regexp.MustCompile(`^".*"$`)

// singleQuotedPattern matches strings wrapped in single quotes.
var singleQuotedPattern = regexp.MustCompile(`^'.*'$`)

// emailPattern provides basic email validation.
// This is lenient - it just checks for the basic structure.
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// ErrInvalidAuthor is returned when an author string is malformed.
var ErrInvalidAuthor = errors.New("invalid author format")

// Author represents the contributor of a change.
//
// Author is lenient in name or email validation, following Copybara's approach.
// An Author can be created from parsing "Name <email>" format or directly
// via NewAuthor.
type Author struct {
	name  string
	email string
}

// Ensure Author implements starlark.Value and starlark.HasAttrs.
var (
	_ starlark.Value    = (*Author)(nil)
	_ starlark.HasAttrs = (*Author)(nil)
)

// NewAuthor creates a new Author with the given name and email.
func NewAuthor(name, email string) *Author {
	return &Author{
		name:  strings.TrimSpace(name),
		email: strings.TrimSpace(email),
	}
}

// ParseAuthor parses an author string in the format "Name <email>".
//
// This is lenient: email can be empty, and it doesn't strictly validate
// that email is an actual email address. Quoted strings are unquoted.
func ParseAuthor(s string) (*Author, error) {
	if s == "" {
		return nil, fmt.Errorf("%w: empty author string", ErrInvalidAuthor)
	}

	// Strip quotes if present
	author := s
	if doubleQuotedPattern.MatchString(s) || singleQuotedPattern.MatchString(s) {
		author = s[1 : len(s)-1]
	}

	matches := authorPattern.FindStringSubmatch(author)
	if matches == nil {
		return nil, fmt.Errorf("%w: '%s' must be in the format 'Name <email>'", ErrInvalidAuthor, s)
	}

	name := strings.TrimSpace(matches[1])
	email := strings.TrimSpace(matches[2])

	if name == "" {
		return nil, fmt.Errorf("%w: author name cannot be empty", ErrInvalidAuthor)
	}

	return &Author{
		name:  name,
		email: email,
	}, nil
}

// ValidateAuthor validates an author string format without creating an Author.
func ValidateAuthor(s string) error {
	_, err := ParseAuthor(s)
	return err
}

// ValidateEmail checks if the email has a valid format.
// Returns nil if valid, error otherwise.
func ValidateEmail(email string) error {
	if email == "" {
		return nil // Empty email is allowed (lenient)
	}
	if !emailPattern.MatchString(email) {
		return fmt.Errorf("invalid email format: %s", email)
	}
	return nil
}

// Name returns the author's name.
func (a *Author) Name() string {
	return a.name
}

// Email returns the author's email address.
func (a *Author) Email() string {
	return a.email
}

// String returns the standard "Name <email>" format.
func (a *Author) String() string {
	return fmt.Sprintf("%s <%s>", a.name, a.email)
}

// Type implements starlark.Value.
func (a *Author) Type() string {
	return "author"
}

// Freeze implements starlark.Value.
func (a *Author) Freeze() {
	// Authors are immutable
}

// Truth implements starlark.Value.
func (a *Author) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (a *Author) Hash() (uint32, error) {
	// Hash based on email if present, otherwise name
	if a.email != "" {
		h := uint32(0)
		for _, c := range a.email {
			h = h*31 + uint32(c)
		}
		return h, nil
	}
	h := uint32(0)
	for _, c := range a.name {
		h = h*31 + uint32(c)
	}
	return h, nil
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
	return []string{"email", "name"}
}

// Equals checks if two authors are equal.
// Authors with the same non-empty email are considered equal.
// If both emails are empty, compare by name.
func (a *Author) Equals(other *Author) bool {
	if other == nil {
		return false
	}
	// If both have non-empty emails, compare by email
	if a.email != "" && other.email != "" {
		return a.email == other.email
	}
	// If both have empty emails, compare by name
	if a.email == "" && other.email == "" {
		return a.name == other.name
	}
	// One has email, one doesn't - not equal
	return false
}

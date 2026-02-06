// Package authoring provides the authoring.* Starlark module for author handling.
//
// The authoring module provides functions for handling commit authoring:
//   - authoring.pass_thru() - Pass through authoring unchanged
//   - authoring.overwrite() - Overwrite with a fixed author
//   - authoring.allowed() - Allow only listed authors
//   - authoring.new_author() - Create an Author from name and email
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/authoring/Authoring.java
package authoring

import (
	"fmt"
	"regexp"
	"slices"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// AuthoringMode defines how authors are mapped from origin to destination.
type AuthoringMode int

const (
	// ModePassThru passes through the original author unchanged.
	ModePassThru AuthoringMode = iota
	// ModeOverwrite always uses the configured default author.
	ModeOverwrite
	// ModeAllowed passes through only allowed authors, using default for others.
	ModeAllowed
)

// String returns the string representation of the mode.
func (m AuthoringMode) String() string {
	switch m {
	case ModePassThru:
		return "PASS_THRU"
	case ModeOverwrite:
		return "OVERWRITE"
	case ModeAllowed:
		return "ALLOWED"
	default:
		return "UNKNOWN"
	}
}

// Module is the authoring.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "authoring",
	Members: starlark.StringDict{
		"pass_thru":  starlark.NewBuiltin("authoring.pass_thru", passThruFn),
		"overwrite":  starlark.NewBuiltin("authoring.overwrite", overwriteFn),
		"allowed":    starlark.NewBuiltin("authoring.allowed", allowedFn),
		"new_author": starlark.NewBuiltin("authoring.new_author", newAuthorFn),
	},
}

// Authoring is the interface implemented by all authoring modes.
type Authoring interface {
	starlark.Value
	// ResolveAuthor resolves the author based on the authoring mode.
	// Takes the original author and returns the resolved author.
	ResolveAuthor(original *Author) *Author
	// DefaultAuthor returns the configured default author.
	DefaultAuthor() *Author
	// Mode returns the authoring mode.
	Mode() AuthoringMode
}

// Ensure all authoring types implement the Authoring interface.
var (
	_ Authoring = (*PassThru)(nil)
	_ Authoring = (*Overwrite)(nil)
	_ Authoring = (*Allowed)(nil)
)

// PassThru represents pass-through authoring.
// The original author is used unchanged.
type PassThru struct {
	defaultAuthor *Author
}

func (p *PassThru) String() string {
	return fmt.Sprintf("authoring.pass_thru(default = %q)", p.defaultAuthor.String())
}
func (p *PassThru) Type() string         { return "authoring.pass_thru" }
func (p *PassThru) Freeze()              {}
func (p *PassThru) Truth() starlark.Bool { return starlark.True }
func (p *PassThru) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: authoring.pass_thru")
}

// ResolveAuthor returns the original author (pass-through mode).
// If original is nil, returns the default author.
func (p *PassThru) ResolveAuthor(original *Author) *Author {
	if original == nil {
		return p.defaultAuthor
	}
	return original
}

// DefaultAuthor returns the configured default author.
func (p *PassThru) DefaultAuthor() *Author {
	return p.defaultAuthor
}

// Mode returns ModePassThru.
func (p *PassThru) Mode() AuthoringMode {
	return ModePassThru
}

// Overwrite represents overwrite authoring.
// The default author is always used, regardless of the original.
type Overwrite struct {
	author *Author
}

func (o *Overwrite) String() string {
	return fmt.Sprintf("authoring.overwrite(default = %q)", o.author.String())
}
func (o *Overwrite) Type() string         { return "authoring.overwrite" }
func (o *Overwrite) Freeze()              {}
func (o *Overwrite) Truth() starlark.Bool { return starlark.True }
func (o *Overwrite) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: authoring.overwrite")
}

// ResolveAuthor always returns the configured author (overwrite mode).
func (o *Overwrite) ResolveAuthor(_ *Author) *Author {
	return o.author
}

// DefaultAuthor returns the configured author.
func (o *Overwrite) DefaultAuthor() *Author {
	return o.author
}

// Mode returns ModeOverwrite.
func (o *Overwrite) Mode() AuthoringMode {
	return ModeOverwrite
}

// Allowed represents allowed authoring with an allowlist.
// Only authors matching the allowlist patterns are passed through.
type Allowed struct {
	defaultAuthor *Author
	allowlist     []string
	patterns      []*regexp.Regexp
}

func (a *Allowed) String() string {
	return fmt.Sprintf("authoring.allowed(default = %q, allowlist = %v)", a.defaultAuthor.String(), a.allowlist)
}
func (a *Allowed) Type() string          { return "authoring.allowed" }
func (a *Allowed) Freeze()               {}
func (a *Allowed) Truth() starlark.Bool  { return starlark.True }
func (a *Allowed) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: authoring.allowed") }

// ResolveAuthor returns the original author if allowed, otherwise the default.
// An author is allowed if their email matches any pattern in the allowlist.
func (a *Allowed) ResolveAuthor(original *Author) *Author {
	if original == nil {
		return a.defaultAuthor
	}
	if a.isAllowed(original) {
		return original
	}
	return a.defaultAuthor
}

// DefaultAuthor returns the configured default author.
func (a *Allowed) DefaultAuthor() *Author {
	return a.defaultAuthor
}

// Mode returns ModeAllowed.
func (a *Allowed) Mode() AuthoringMode {
	return ModeAllowed
}

// Allowlist returns the list of allowed patterns.
func (a *Allowed) Allowlist() []string {
	return slices.Clone(a.allowlist)
}

// isAllowed checks if the author is in the allowlist.
// Checks both email and author string against patterns.
func (a *Allowed) isAllowed(author *Author) bool {
	email := author.Email()
	authorStr := author.String()

	for _, pattern := range a.allowlist {
		// Exact match on email
		if pattern == email {
			return true
		}
		// Exact match on full author string
		if pattern == authorStr {
			return true
		}
	}

	// Check regex patterns
	for _, re := range a.patterns {
		if re.MatchString(email) || re.MatchString(authorStr) {
			return true
		}
	}

	return false
}

// passThruFn implements authoring.pass_thru().
//
// Parameters:
//   - default (required): The default author string in "Name <email>" format.
//     Used for squash mode workflows or when author cannot be determined.
func passThruFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var defaultAuthor string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"default", &defaultAuthor,
	); err != nil {
		return nil, err
	}

	author, err := ParseAuthor(defaultAuthor)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn.Name(), err)
	}

	return &PassThru{defaultAuthor: author}, nil
}

// overwriteFn implements authoring.overwrite().
//
// Parameters:
//   - default (required): The author to use for all commits in "Name <email>" format.
func overwriteFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var author string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"default", &author,
	); err != nil {
		return nil, err
	}

	parsedAuthor, err := ParseAuthor(author)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn.Name(), err)
	}

	return &Overwrite{author: parsedAuthor}, nil
}

// allowedFn implements authoring.allowed().
//
// Parameters:
//   - default (required): The default author string in "Name <email>" format.
//     Used for squash mode workflows or when author is not in allowlist.
//   - allowlist (required): List of allowed author patterns.
//     Can be exact emails, full author strings, or regex patterns.
func allowedFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var defaultAuthor string
	var allowlist *starlark.List

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"default", &defaultAuthor,
		"allowlist", &allowlist,
	); err != nil {
		return nil, err
	}

	// Parse and validate default author
	author, err := ParseAuthor(defaultAuthor)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn.Name(), err)
	}

	// Validate allowlist is not empty
	if allowlist == nil || allowlist.Len() == 0 {
		return nil, fmt.Errorf("%s: 'allowlist' must be a non-empty list. For default mapping, use 'overwrite(...)' mode instead", fn.Name())
	}

	// Extract and validate allowlist entries
	patterns := make([]string, 0, allowlist.Len())
	seen := make(map[string]bool)

	for i := range allowlist.Len() {
		s, ok := allowlist.Index(i).(starlark.String)
		if !ok {
			return nil, fmt.Errorf("%s: allowlist entry %d must be a string, got %s", fn.Name(), i, allowlist.Index(i).Type())
		}

		pattern := string(s)
		if pattern == "" {
			return nil, fmt.Errorf("%s: allowlist entry %d cannot be empty", fn.Name(), i)
		}

		if seen[pattern] {
			return nil, fmt.Errorf("%s: duplicate allowlist entry '%s'", fn.Name(), pattern)
		}
		seen[pattern] = true
		patterns = append(patterns, pattern)
	}

	// Compile regex patterns (patterns starting and ending with /)
	var regexPatterns []*regexp.Regexp
	for _, p := range patterns {
		if len(p) >= 2 && p[0] == '/' && p[len(p)-1] == '/' {
			// This is a regex pattern
			re, err := regexp.Compile(p[1 : len(p)-1])
			if err != nil {
				return nil, fmt.Errorf("%s: invalid regex pattern '%s': %w", fn.Name(), p, err)
			}
			regexPatterns = append(regexPatterns, re)
		}
	}

	return &Allowed{
		defaultAuthor: author,
		allowlist:     patterns,
		patterns:      regexPatterns,
	}, nil
}

// newAuthorFn implements authoring.new_author().
//
// Parameters:
//   - name (required): The author's name.
//   - email (required): The author's email address.
func newAuthorFn(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, email string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"email", &email,
	); err != nil {
		return nil, err
	}

	if name == "" {
		return nil, fmt.Errorf("%s: name cannot be empty", fn.Name())
	}

	return NewAuthor(name, email), nil
}

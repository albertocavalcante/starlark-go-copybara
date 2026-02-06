package metadata

import (
	"fmt"

	"go.starlark.net/starlark"
)

// SquashNotes is a transformation that squashes commit notes.
type SquashNotes struct {
	prefix          string
	compact         bool
	showRef         bool
	showAuthor      bool
	showDescription bool
	oldestFirst     bool
}

func (s *SquashNotes) String() string        { return "metadata.squash_notes()" }
func (s *SquashNotes) Type() string          { return "squash_notes" }
func (s *SquashNotes) Freeze()               {}
func (s *SquashNotes) Truth() starlark.Bool  { return starlark.True }
func (s *SquashNotes) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: squash_notes") }

// SaveAuthor is a transformation that saves the original author.
type SaveAuthor struct {
	label string
}

func (s *SaveAuthor) String() string        { return fmt.Sprintf("metadata.save_author(label = %q)", s.label) }
func (s *SaveAuthor) Type() string          { return "save_author" }
func (s *SaveAuthor) Freeze()               {}
func (s *SaveAuthor) Truth() starlark.Bool  { return starlark.True }
func (s *SaveAuthor) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: save_author") }

// ReplaceMessage is a transformation that replaces the commit message.
type ReplaceMessage struct {
	message string
}

func (r *ReplaceMessage) String() string        { return fmt.Sprintf("metadata.replace_message(%q)", r.message) }
func (r *ReplaceMessage) Type() string          { return "replace_message" }
func (r *ReplaceMessage) Freeze()               {}
func (r *ReplaceMessage) Truth() starlark.Bool  { return starlark.True }
func (r *ReplaceMessage) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: replace_message") }

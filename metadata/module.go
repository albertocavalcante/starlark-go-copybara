// Package metadata provides the metadata.* Starlark module for commit message transformations.
//
// The metadata module provides functions for transforming commit messages:
//   - metadata.squash_notes() - Squash commit messages
//   - metadata.save_author() - Preserve original author
//   - metadata.replace_message() - Transform commit message
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/MetadataModule.java
package metadata

import (
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the metadata.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "metadata",
	Members: starlark.StringDict{
		"squash_notes":    starlark.NewBuiltin("metadata.squash_notes", squashNotesFn),
		"save_author":     starlark.NewBuiltin("metadata.save_author", saveAuthorFn),
		"replace_message": starlark.NewBuiltin("metadata.replace_message", replaceMessageFn),
	},
}

// squashNotesFn implements metadata.squash_notes().
func squashNotesFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		prefix       string
		compact      bool
		showRef      bool
		showAuthor   bool
		showDescription bool
		oldest_first bool
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"prefix?", &prefix,
		"compact?", &compact,
		"show_ref?", &showRef,
		"show_author?", &showAuthor,
		"show_description?", &showDescription,
		"oldest_first?", &oldest_first,
	); err != nil {
		return nil, err
	}

	return &SquashNotes{
		prefix:          prefix,
		compact:         compact,
		showRef:         showRef,
		showAuthor:      showAuthor,
		showDescription: showDescription,
		oldestFirst:     oldest_first,
	}, nil
}

// saveAuthorFn implements metadata.save_author().
func saveAuthorFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var label string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"label?", &label,
	); err != nil {
		return nil, err
	}

	if label == "" {
		label = "ORIGINAL_AUTHOR"
	}

	return &SaveAuthor{label: label}, nil
}

// replaceMessageFn implements metadata.replace_message().
func replaceMessageFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var message string

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"message", &message,
	); err != nil {
		return nil, err
	}

	return &ReplaceMessage{message: message}, nil
}

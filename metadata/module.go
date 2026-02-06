// Package metadata provides the metadata.* Starlark module for commit message transformations.
//
// The metadata module provides functions for transforming commit messages:
//   - metadata.squash_notes() - Squash commit messages
//   - metadata.save_author() - Preserve original author
//   - metadata.restore_author() - Restore author from a label
//   - metadata.replace_message() - Transform commit message
//   - metadata.expose_label() - Expose a label from the commit message
//   - metadata.add_header() - Add a header to the commit message
//   - metadata.scrubber() - Scrub sensitive content from commit messages
//   - metadata.map_author() - Map author identities
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/MetadataModule.java
package metadata

import (
	"fmt"
	"regexp"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the metadata.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "metadata",
	Members: starlark.StringDict{
		"squash_notes":    starlark.NewBuiltin("metadata.squash_notes", squashNotesFn),
		"save_author":     starlark.NewBuiltin("metadata.save_author", saveAuthorFn),
		"restore_author":  starlark.NewBuiltin("metadata.restore_author", restoreAuthorFn),
		"replace_message": starlark.NewBuiltin("metadata.replace_message", replaceMessageFn),
		"expose_label":    starlark.NewBuiltin("metadata.expose_label", exposeLabelFn),
		"add_header":      starlark.NewBuiltin("metadata.add_header", addHeaderFn),
		"scrubber":        starlark.NewBuiltin("metadata.scrubber", scrubberFn),
		"map_author":      starlark.NewBuiltin("metadata.map_author", mapAuthorFn),
	},
}

// squashNotesFn implements metadata.squash_notes().
//
// Reference: MetadataModule.java squashNotes()
func squashNotesFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		prefix          string = "Copybara import of the project:\n\n"
		max             int    = 100
		compact         bool   = true
		showRef         bool   = true
		showAuthor      bool   = true
		showDescription bool   = true
		oldestFirst     bool   = false
		useMerge        bool   = true
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"prefix?", &prefix,
		"max?", &max,
		"compact?", &compact,
		"show_ref?", &showRef,
		"show_author?", &showAuthor,
		"show_description?", &showDescription,
		"oldest_first?", &oldestFirst,
		"use_merge?", &useMerge,
	); err != nil {
		return nil, err
	}

	if prefix == "" {
		return nil, fmt.Errorf("prefix cannot be empty")
	}

	return &SquashNotes{
		prefix:          prefix,
		max:             max,
		compact:         compact,
		showRef:         showRef,
		showAuthor:      showAuthor,
		showDescription: showDescription,
		oldestFirst:     oldestFirst,
		useMerge:        useMerge,
	}, nil
}

// saveAuthorFn implements metadata.save_author().
//
// Reference: MetadataModule.java saveAuthor()
func saveAuthorFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		label     string = "ORIGINAL_AUTHOR"
		separator string = "="
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"label?", &label,
		"separator?", &separator,
	); err != nil {
		return nil, err
	}

	return &SaveAuthor{
		label:     label,
		separator: separator,
	}, nil
}

// restoreAuthorFn implements metadata.restore_author().
//
// Reference: MetadataModule.java restoreAuthor()
func restoreAuthorFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		label            string = "ORIGINAL_AUTHOR"
		separator        string = "="
		searchAllChanges bool   = false
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"label?", &label,
		"separator?", &separator,
		"search_all_changes?", &searchAllChanges,
	); err != nil {
		return nil, err
	}

	return &RestoreAuthor{
		label:            label,
		separator:        separator,
		searchAllChanges: searchAllChanges,
	}, nil
}

// replaceMessageFn implements metadata.replace_message().
//
// Reference: MetadataModule.java replaceMessage()
func replaceMessageFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		text                string
		ignoreLabelNotFound bool = false
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"text", &text,
		"ignore_label_not_found?", &ignoreLabelNotFound,
	); err != nil {
		return nil, err
	}

	return &ReplaceMessage{
		message:             text,
		ignoreLabelNotFound: ignoreLabelNotFound,
	}, nil
}

// exposeLabelFn implements metadata.expose_label().
//
// Reference: MetadataModule.java exposeLabel()
func exposeLabelFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		name                string
		newName             starlark.Value = starlark.None
		separator           string         = "="
		ignoreLabelNotFound bool           = true
		all                 bool           = false
		concatSeparator     starlark.Value = starlark.None
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"new_name?", &newName,
		"separator?", &separator,
		"ignore_label_not_found?", &ignoreLabelNotFound,
		"all?", &all,
		"concat_separator?", &concatSeparator,
	); err != nil {
		return nil, err
	}

	// Validate label name
	if !isValidLabelName(name) {
		return nil, fmt.Errorf("'name': Invalid label name '%s'", name)
	}

	// Get new name, defaulting to the original name
	actualNewName := name
	if newName != starlark.None {
		s, ok := starlark.AsString(newName)
		if !ok {
			return nil, fmt.Errorf("new_name must be a string")
		}
		actualNewName = s
		if !isValidLabelName(actualNewName) {
			return nil, fmt.Errorf("'new_name': Invalid label name '%s'", actualNewName)
		}
	}

	// Get concat separator if provided
	var concatSep string
	hasConcatSep := false
	if concatSeparator != starlark.None {
		if !all {
			return nil, fmt.Errorf("'concat_separator': Cannot be set unless all is True")
		}
		s, ok := starlark.AsString(concatSeparator)
		if !ok {
			return nil, fmt.Errorf("concat_separator must be a string")
		}
		concatSep = s
		hasConcatSep = true
	}

	return &ExposeLabel{
		name:                name,
		newName:             actualNewName,
		separator:           separator,
		ignoreLabelNotFound: ignoreLabelNotFound,
		all:                 all,
		concatSeparator:     concatSep,
		hasConcatSeparator:  hasConcatSep,
	}, nil
}

// addHeaderFn implements metadata.add_header().
//
// Reference: MetadataModule.java addHeader()
func addHeaderFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		text                string
		ignoreLabelNotFound bool = false
		newLine             bool = true
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"text", &text,
		"ignore_label_not_found?", &ignoreLabelNotFound,
		"new_line?", &newLine,
	); err != nil {
		return nil, err
	}

	return &AddHeader{
		text:                text,
		ignoreLabelNotFound: ignoreLabelNotFound,
		newLine:             newLine,
	}, nil
}

// scrubberFn implements metadata.scrubber().
//
// Reference: MetadataModule.java scrubber()
func scrubberFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		regex         string
		regexes       *starlark.List
		msgIfNoMatch  starlark.Value = starlark.None
		failIfNoMatch bool           = false
		replacement   string         = ""
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"regex?", &regex,
		"regexes?", &regexes,
		"msg_if_no_match?", &msgIfNoMatch,
		"fail_if_no_match?", &failIfNoMatch,
		"replacement?", &replacement,
	); err != nil {
		return nil, err
	}

	// Build list of regexes
	var patterns []*regexp.Regexp

	if regex != "" {
		re, err := regexp.Compile("(?m)" + regex) // Multiline mode
		if err != nil {
			return nil, fmt.Errorf("invalid regex expression: %v", err)
		}
		patterns = append(patterns, re)
	}

	if regexes != nil {
		for i := 0; i < regexes.Len(); i++ {
			item := regexes.Index(i)
			s, ok := starlark.AsString(item)
			if !ok {
				return nil, fmt.Errorf("regexes must be a list of strings")
			}
			re, err := regexp.Compile("(?m)" + s) // Multiline mode
			if err != nil {
				return nil, fmt.Errorf("invalid regex expression at index %d: %v", i, err)
			}
			patterns = append(patterns, re)
		}
	}

	if len(patterns) == 0 {
		return nil, fmt.Errorf("at least one regex must be provided via 'regex' or 'regexes'")
	}

	// Get msg_if_no_match
	var msgNoMatch string
	if msgIfNoMatch != starlark.None {
		s, ok := starlark.AsString(msgIfNoMatch)
		if !ok {
			return nil, fmt.Errorf("msg_if_no_match must be a string")
		}
		msgNoMatch = s
	}

	// Validate fail_if_no_match and msg_if_no_match combination
	if failIfNoMatch && msgNoMatch != "" {
		return nil, fmt.Errorf("if fail_if_no_match is true, msg_if_no_match should be None")
	}

	return &Scrubber{
		regexes:       patterns,
		replacement:   replacement,
		msgIfNoMatch:  msgNoMatch,
		failIfNoMatch: failIfNoMatch,
	}, nil
}

// mapAuthorFn implements metadata.map_author().
//
// Reference: MetadataModule.java mapAuthor()
func mapAuthorFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		authors               *starlark.Dict
		reversible            bool = false
		noopReverse           bool = false
		failIfNotFound        bool = false
		reverseFailIfNotFound bool = false
		mapAllChanges         bool = false
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"authors", &authors,
		"reversible?", &reversible,
		"noop_reverse?", &noopReverse,
		"fail_if_not_found?", &failIfNotFound,
		"reverse_fail_if_not_found?", &reverseFailIfNotFound,
		"map_all_changes?", &mapAllChanges,
	); err != nil {
		return nil, err
	}

	// Validate parameter combinations
	if reverseFailIfNotFound && !reversible {
		return nil, fmt.Errorf("'reverse_fail_if_not_found' can only be true if 'reversible' is true")
	}

	if reverseFailIfNotFound && noopReverse {
		return nil, fmt.Errorf("'reverse_fail_if_not_found' can only be true if 'noop_reverse' is not set")
	}

	// Convert authors dict to Go map
	authorMap := make(map[string]string)
	for _, item := range authors.Items() {
		key, ok := starlark.AsString(item[0])
		if !ok {
			return nil, fmt.Errorf("authors keys must be strings")
		}
		value, ok := starlark.AsString(item[1])
		if !ok {
			return nil, fmt.Errorf("authors values must be strings")
		}
		authorMap[key] = value
	}

	return NewMapAuthor(authorMap, reversible, noopReverse, failIfNotFound, reverseFailIfNotFound, mapAllChanges)
}

// isValidLabelName checks if a label name is valid.
// Label names must start with a letter and contain only letters, digits, underscores, and hyphens.
var validLabelPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func isValidLabelName(name string) bool {
	return validLabelPattern.MatchString(name)
}

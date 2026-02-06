// Package transform provides the transformation execution context.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/TransformWork.java
package transform

import (
	"io/fs"
	"regexp"
	"strings"
)

// Change represents a single change in a migration.
type Change struct {
	// Ref is the change reference (e.g., commit hash).
	Ref string

	// Author is the change author.
	Author string

	// Message is the commit message.
	Message string

	// MappedAuthor is the mapped author (after transformations).
	MappedAuthor string

	// IsMerge indicates if this is a merge commit.
	IsMerge bool

	// Labels contains key-value labels from the change.
	Labels map[string][]string

	// Files is the list of changed files (optional).
	Files []string
}

// FirstLineMessage returns the first line of the commit message.
func (c *Change) FirstLineMessage() string {
	if idx := strings.Index(c.Message, "\n"); idx != -1 {
		return c.Message[:idx]
	}
	return c.Message
}

// Changes holds the list of changes being migrated.
type Changes struct {
	// Current is the list of changes in the current migration.
	Current []*Change
}

// Context provides the execution context for transformations.
type Context struct {
	// WorkDir is the working directory for file operations.
	WorkDir string

	// FS is the filesystem for file operations.
	FS fs.FS

	// DryRun indicates whether to simulate without making changes.
	DryRun bool

	// Message is the commit message being transformed.
	Message string

	// Author is the commit author.
	Author string

	// Changes contains the changes being migrated.
	Changes *Changes

	// Labels contains message labels extracted from the commit message.
	Labels map[string][]string
}

// NewContext creates a new transformation context.
func NewContext(workDir string) *Context {
	return &Context{
		WorkDir: workDir,
		Labels:  make(map[string][]string),
		Changes: &Changes{},
	}
}

// labelPattern matches labels in the format "Label-Name: value" or "Label-Name=value".
var labelPattern = regexp.MustCompile(`(?m)^([A-Za-z][A-Za-z0-9_-]*)[:=]\s*(.+)$`)

// GetLabel returns the first value of a label from the message.
func (ctx *Context) GetLabel(name string) string {
	// First check pre-populated labels
	if values, ok := ctx.Labels[name]; ok && len(values) > 0 {
		return values[0]
	}

	// Then search in message
	matches := labelPattern.FindAllStringSubmatch(ctx.Message, -1)
	for _, match := range matches {
		if match[1] == name {
			return strings.TrimSpace(match[2])
		}
	}

	// Search in changes
	if ctx.Changes != nil {
		for _, change := range ctx.Changes.Current {
			if change.Labels != nil {
				if values, ok := change.Labels[name]; ok && len(values) > 0 {
					return values[0]
				}
			}
		}
	}

	return ""
}

// GetAllLabels returns all values of a label from the message and changes.
func (ctx *Context) GetAllLabels(name string) []string {
	seen := make(map[string]bool)
	var values []string

	addValue := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" && !seen[v] {
			seen[v] = true
			values = append(values, v)
		}
	}

	// First check pre-populated labels
	if labelValues, ok := ctx.Labels[name]; ok {
		for _, v := range labelValues {
			addValue(v)
		}
	}

	// Then search in message
	matches := labelPattern.FindAllStringSubmatch(ctx.Message, -1)
	for _, match := range matches {
		if match[1] == name {
			addValue(match[2])
		}
	}

	// Search in changes
	if ctx.Changes != nil {
		for _, change := range ctx.Changes.Current {
			if change.Labels != nil {
				if labelValues, ok := change.Labels[name]; ok {
					for _, v := range labelValues {
						addValue(v)
					}
				}
			}
		}
	}

	return values
}

// AddLabel adds a label to the message.
func (ctx *Context) AddLabel(name, value, separator string) {
	// Add to labels map
	ctx.Labels[name] = append(ctx.Labels[name], value)

	// Add to message
	line := name + separator + value
	if ctx.Message != "" && !strings.HasSuffix(ctx.Message, "\n") {
		ctx.Message += "\n"
	}
	ctx.Message += line + "\n"
}

// RemoveLabel removes a label from the message.
func (ctx *Context) RemoveLabel(name string) {
	// Remove from labels map
	delete(ctx.Labels, name)

	// Remove from message using a pattern that matches the label
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `[:=].*\n?`)
	ctx.Message = pattern.ReplaceAllString(ctx.Message, "")
}

// RemoveLabelWithValue removes a specific label-value pair from the message.
func (ctx *Context) RemoveLabelWithValue(name, value string) {
	// Remove from labels map
	if values, ok := ctx.Labels[name]; ok {
		var newValues []string
		for _, v := range values {
			if v != value {
				newValues = append(newValues, v)
			}
		}
		if len(newValues) > 0 {
			ctx.Labels[name] = newValues
		} else {
			delete(ctx.Labels, name)
		}
	}

	// Remove from message
	escapedValue := regexp.QuoteMeta(value)
	pattern := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `[:=]\s*` + escapedValue + `\s*\n?`)
	ctx.Message = pattern.ReplaceAllString(ctx.Message, "")
}

// Result contains the result of a transformation.
type Result struct {
	// FilesChanged lists the files that were modified.
	FilesChanged []string

	// Summary describes what the transformation did.
	Summary string
}

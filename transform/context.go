// Package transform provides the transformation execution context.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/TransformWork.java
package transform

import (
	"io/fs"
)

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
}

// NewContext creates a new transformation context.
func NewContext(workDir string) *Context {
	return &Context{
		WorkDir: workDir,
	}
}

// Result contains the result of a transformation.
type Result struct {
	// FilesChanged lists the files that were modified.
	FilesChanged []string

	// Summary describes what the transformation did.
	Summary string
}

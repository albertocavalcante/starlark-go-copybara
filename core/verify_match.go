package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// VerifyMatch is a pseudo-transformation that verifies files match a regex.
//
// It does not transform any code, but will return errors on failure.
// Not applied in reversals unless also_on_reversal is set.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/VerifyMatch.java
type VerifyMatch struct {
	regex          *regexp.Regexp
	regexStr       string
	paths          *Glob
	verifyNoMatch  bool
	alsoOnReversal bool
	failureMessage string
}

var _ Transformation = (*VerifyMatch)(nil)

// String implements starlark.Value.
func (v *VerifyMatch) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("regex = %q", v.regexStr))

	if v.paths != nil && !v.paths.IsAllFiles() {
		parts = append(parts, fmt.Sprintf("paths = %s", v.paths))
	}

	if v.verifyNoMatch {
		parts = append(parts, "verify_no_match = True")
	}

	if v.alsoOnReversal {
		parts = append(parts, "also_on_reversal = True")
	}

	if v.failureMessage != "" {
		parts = append(parts, fmt.Sprintf("failure_message = %q", v.failureMessage))
	}

	return fmt.Sprintf("core.verify_match(%s)", strings.Join(parts, ", "))
}

// Type implements starlark.Value.
func (v *VerifyMatch) Type() string {
	return "verify_match"
}

// Freeze implements starlark.Value.
func (v *VerifyMatch) Freeze() {}

// Truth implements starlark.Value.
func (v *VerifyMatch) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (v *VerifyMatch) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: verify_match")
}

// Apply implements Transformation.
func (v *VerifyMatch) Apply(ctx *transform.Context) error {
	if ctx.WorkDir == "" {
		return fmt.Errorf("workdir is required for verify_match transformation")
	}

	var errors []string

	// Walk the workdir and check files matching the glob
	err := filepath.WalkDir(ctx.WorkDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and symlinks
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		relPath, err := filepath.Rel(ctx.WorkDir, path)
		if err != nil {
			return err
		}

		// Check if file matches glob
		if !v.paths.Matches(relPath) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path) //nolint:gosec // path is from WalkDir in workdir
		if err != nil {
			return fmt.Errorf("failed to read file %q: %w", relPath, err)
		}

		// Check regex match
		matches := v.regex.FindIndex(content)
		hasMatch := matches != nil

		if v.verifyNoMatch && hasMatch {
			// Found match when we expected no match
			matchStr := string(content[matches[0]:matches[1]])
			line := countLines(content[:matches[0]]) + 1
			errMsg := fmt.Sprintf("%s - Unexpected match found at line %d - '%s'",
				relPath, line, truncate(matchStr, 50))
			if v.failureMessage != "" {
				errMsg += "\n" + v.failureMessage
			}
			errors = append(errors, errMsg)
		} else if !v.verifyNoMatch && !hasMatch {
			// Expected match but found none
			errMsg := fmt.Sprintf("%s - Expected string was not present", relPath)
			if v.failureMessage != "" {
				errMsg += "\n" + v.failureMessage
			}
			errors = append(errors, errMsg)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(errors) > 0 {
		return &VerifyMatchError{
			Errors:      errors,
			Description: v.Describe(),
		}
	}

	return nil
}

// countLines counts the number of newlines in a byte slice.
func countLines(data []byte) int {
	count := 0
	for _, b := range data {
		if b == '\n' {
			count++
		}
	}
	return count
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// VerifyMatchError represents a verification failure.
type VerifyMatchError struct {
	Errors      []string
	Description string
}

func (e *VerifyMatchError) Error() string {
	return fmt.Sprintf("%d file(s) failed the validation of %s:\n%s",
		len(e.Errors), e.Description, strings.Join(e.Errors, "\n"))
}

// Reverse implements Transformation.
func (v *VerifyMatch) Reverse() transform.Transformation {
	if v.alsoOnReversal {
		return v
	}
	return transform.NewNoopTransformation(v)
}

// Describe implements Transformation.
func (v *VerifyMatch) Describe() string {
	if v.verifyNoMatch {
		return fmt.Sprintf("verify_no_match '%s'", v.regexStr)
	}
	return fmt.Sprintf("verify_match '%s'", v.regexStr)
}

// Regex returns the compiled regex.
func (v *VerifyMatch) Regex() *regexp.Regexp {
	return v.regex
}

// RegexString returns the original regex string.
func (v *VerifyMatch) RegexString() string {
	return v.regexStr
}

// Paths returns the glob filter.
func (v *VerifyMatch) Paths() *Glob {
	return v.paths
}

// VerifyNoMatch returns whether this verifies no match.
func (v *VerifyMatch) VerifyNoMatch() bool {
	return v.verifyNoMatch
}

// AlsoOnReversal returns whether this applies on reversal.
func (v *VerifyMatch) AlsoOnReversal() bool {
	return v.alsoOnReversal
}

// FailureMessage returns the custom failure message.
func (v *VerifyMatch) FailureMessage() string {
	return v.failureMessage
}

// verifyMatchFn implements core.verify_match().
func verifyMatchFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		regexStr       string
		paths          starlark.Value = starlark.None
		verifyNoMatch  bool
		alsoOnReversal bool
		failureMessage starlark.Value = starlark.None
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"regex", &regexStr,
		"paths?", &paths,
		"verify_no_match?", &verifyNoMatch,
		"also_on_reversal?", &alsoOnReversal,
		"failure_message?", &failureMessage,
	); err != nil {
		return nil, err
	}

	// Compile regex with multiline mode
	regex, err := regexp.Compile("(?m)" + regexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid regex %q: %w", regexStr, err)
	}

	vm := &VerifyMatch{
		regex:          regex,
		regexStr:       regexStr,
		verifyNoMatch:  verifyNoMatch,
		alsoOnReversal: alsoOnReversal,
	}

	// Handle paths parameter
	switch v := paths.(type) {
	case starlark.NoneType:
		vm.paths = AllFiles()
	case *Glob:
		vm.paths = v
	case *starlark.List:
		patterns := make([]string, v.Len())
		for i := range v.Len() {
			s, ok := starlark.AsString(v.Index(i))
			if !ok {
				return nil, fmt.Errorf("paths must be strings, got %s", v.Index(i).Type())
			}
			patterns[i] = s
		}
		vm.paths, err = NewGlob(patterns, nil)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("paths must be a glob or list of strings, got %s", paths.Type())
	}

	// Handle failure_message parameter
	switch v := failureMessage.(type) {
	case starlark.NoneType:
		// keep empty
	case starlark.String:
		vm.failureMessage = string(v)
	default:
		return nil, fmt.Errorf("failure_message must be a string, got %s", failureMessage.Type())
	}

	return vm, nil
}

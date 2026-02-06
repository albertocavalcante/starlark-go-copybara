package git

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// IntegrateStrategy defines the integration strategy.
type IntegrateStrategy string

const (
	// StrategyFakeMerge adds the url revision as parent but ignores files.
	StrategyFakeMerge IntegrateStrategy = "FAKE_MERGE"
	// StrategyFakeMergeAndIncludeFiles is like FAKE_MERGE but includes non-destination files.
	StrategyFakeMergeAndIncludeFiles IntegrateStrategy = "FAKE_MERGE_AND_INCLUDE_FILES"
	// StrategyIncludeFiles includes non-destination files without creating a merge.
	StrategyIncludeFiles IntegrateStrategy = "INCLUDE_FILES"
)

// ParseIntegrateStrategy parses a string into an IntegrateStrategy.
func ParseIntegrateStrategy(s string) (IntegrateStrategy, error) {
	switch strings.ToUpper(s) {
	case "FAKE_MERGE":
		return StrategyFakeMerge, nil
	case "FAKE_MERGE_AND_INCLUDE_FILES", "":
		return StrategyFakeMergeAndIncludeFiles, nil
	case "INCLUDE_FILES":
		return StrategyIncludeFiles, nil
	default:
		return "", fmt.Errorf("invalid integrate strategy: %q", s)
	}
}

// DefaultIntegrateLabel is the default label for integration.
const DefaultIntegrateLabel = "COPYBARA_INTEGRATE_REVIEW"

// IntegrateChanges represents git.integrate() configuration.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitIntegrateChanges.java
type IntegrateChanges struct {
	label        string
	strategy     IntegrateStrategy
	ignoreErrors bool
}

// NewIntegrateChanges creates a new IntegrateChanges.
func NewIntegrateChanges(label string, strategy IntegrateStrategy, ignoreErrors bool) *IntegrateChanges {
	if label == "" {
		label = DefaultIntegrateLabel
	}
	return &IntegrateChanges{
		label:        label,
		strategy:     strategy,
		ignoreErrors: ignoreErrors,
	}
}

// String implements starlark.Value.
func (i *IntegrateChanges) String() string {
	var parts []string
	if i.label != DefaultIntegrateLabel {
		parts = append(parts, fmt.Sprintf("label = %q", i.label))
	}
	if i.strategy != StrategyFakeMergeAndIncludeFiles {
		parts = append(parts, fmt.Sprintf("strategy = %q", i.strategy))
	}
	if !i.ignoreErrors {
		parts = append(parts, "ignore_errors = False")
	}
	if len(parts) == 0 {
		return "git.integrate()"
	}
	return fmt.Sprintf("git.integrate(%s)", strings.Join(parts, ", "))
}

// Type implements starlark.Value.
func (i *IntegrateChanges) Type() string {
	return "git.integrate"
}

// Freeze implements starlark.Value.
func (i *IntegrateChanges) Freeze() {}

// Truth implements starlark.Value.
func (i *IntegrateChanges) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (i *IntegrateChanges) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.integrate")
}

// Label returns the integration label.
func (i *IntegrateChanges) Label() string {
	return i.label
}

// Strategy returns the integration strategy.
func (i *IntegrateChanges) Strategy() IntegrateStrategy {
	return i.strategy
}

// IgnoreErrors returns whether to ignore integration errors.
func (i *IntegrateChanges) IgnoreErrors() bool {
	return i.ignoreErrors
}

// Attr implements starlark.HasAttrs.
func (i *IntegrateChanges) Attr(name string) (starlark.Value, error) {
	switch name {
	case "label":
		return starlark.String(i.label), nil
	case "strategy":
		return starlark.String(i.strategy), nil
	case "ignore_errors":
		return starlark.Bool(i.ignoreErrors), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (i *IntegrateChanges) AttrNames() []string {
	return []string{"label", "strategy", "ignore_errors"}
}

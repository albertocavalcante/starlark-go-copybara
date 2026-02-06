package git

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// ReviewState defines the required review state for GitHub PRs.
type ReviewState string

const (
	// ReviewStateHeadCommitApproved requires the head commit to be approved.
	ReviewStateHeadCommitApproved ReviewState = "HEAD_COMMIT_APPROVED"
	// ReviewStateAnyCommitApproved accepts any approved commit.
	ReviewStateAnyCommitApproved ReviewState = "ANY_COMMIT_APPROVED"
	// ReviewStateHasReviewers requires reviewers to be present.
	ReviewStateHasReviewers ReviewState = "HAS_REVIEWERS"
	// ReviewStateAny accepts any review state.
	ReviewStateAny ReviewState = "ANY"
)

// ParseReviewState parses a string into a ReviewState.
func ParseReviewState(s string) (ReviewState, error) {
	switch strings.ToUpper(s) {
	case "HEAD_COMMIT_APPROVED":
		return ReviewStateHeadCommitApproved, nil
	case "ANY_COMMIT_APPROVED":
		return ReviewStateAnyCommitApproved, nil
	case "HAS_REVIEWERS":
		return ReviewStateHasReviewers, nil
	case "ANY", "":
		return ReviewStateAny, nil
	default:
		return "", fmt.Errorf("invalid review state: %q", s)
	}
}

// StateFilter defines the PR state filter.
type StateFilter string

const (
	// StateFilterOpen only migrates open PRs.
	StateFilterOpen StateFilter = "OPEN"
	// StateFilterClosed only migrates closed PRs.
	StateFilterClosed StateFilter = "CLOSED"
	// StateFilterAll migrates all PRs.
	StateFilterAll StateFilter = "ALL"
)

// ParseStateFilter parses a string into a StateFilter.
func ParseStateFilter(s string) (StateFilter, error) {
	switch strings.ToUpper(s) {
	case "OPEN", "":
		return StateFilterOpen, nil
	case "CLOSED":
		return StateFilterClosed, nil
	case "ALL":
		return StateFilterAll, nil
	default:
		return "", fmt.Errorf("invalid state filter: %q", s)
	}
}

// GitHubOrigin represents a GitHub repository origin.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitHubPrOrigin.java
type GitHubOrigin struct {
	url                    string
	ref                    string
	branch                 string
	submodules             SubmoduleStrategy
	firstParent            bool
	partialFetch           bool
	state                  StateFilter
	reviewState            ReviewState
	primaryBranchMigration bool
}

// String implements starlark.Value.
func (g *GitHubOrigin) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("url = %q", g.url))
	if g.ref != "" {
		parts = append(parts, fmt.Sprintf("ref = %q", g.ref))
	}
	if g.branch != "" {
		parts = append(parts, fmt.Sprintf("branch = %q", g.branch))
	}
	if g.reviewState != "" && g.reviewState != ReviewStateAny {
		parts = append(parts, fmt.Sprintf("review_state = %q", g.reviewState))
	}
	if g.state != "" && g.state != StateFilterOpen {
		parts = append(parts, fmt.Sprintf("state = %q", g.state))
	}
	return fmt.Sprintf("git.github_origin(%s)", strings.Join(parts, ", "))
}

// Type implements starlark.Value.
func (g *GitHubOrigin) Type() string {
	return "git.github_origin"
}

// Freeze implements starlark.Value.
func (g *GitHubOrigin) Freeze() {}

// Truth implements starlark.Value.
func (g *GitHubOrigin) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (g *GitHubOrigin) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.github_origin")
}

// URL returns the repository URL.
func (g *GitHubOrigin) URL() string {
	return g.url
}

// Ref returns the Git reference.
func (g *GitHubOrigin) Ref() string {
	return g.ref
}

// Branch returns the branch filter.
func (g *GitHubOrigin) Branch() string {
	return g.branch
}

// Submodules returns the submodule strategy.
func (g *GitHubOrigin) Submodules() SubmoduleStrategy {
	return g.submodules
}

// FirstParent returns whether to use first parent only.
func (g *GitHubOrigin) FirstParent() bool {
	return g.firstParent
}

// PartialFetch returns whether partial fetch is enabled.
func (g *GitHubOrigin) PartialFetch() bool {
	return g.partialFetch
}

// State returns the PR state filter.
func (g *GitHubOrigin) State() StateFilter {
	return g.state
}

// ReviewState returns the required review state.
func (g *GitHubOrigin) ReviewState() ReviewState {
	return g.reviewState
}

// PrimaryBranchMigration returns whether primary branch migration mode is enabled.
func (g *GitHubOrigin) PrimaryBranchMigration() bool {
	return g.primaryBranchMigration
}

// Attr implements starlark.HasAttrs.
func (g *GitHubOrigin) Attr(name string) (starlark.Value, error) {
	switch name {
	case "url":
		return starlark.String(g.url), nil
	case "ref":
		return starlark.String(g.ref), nil
	case "branch":
		return starlark.String(g.branch), nil
	case "submodules":
		return starlark.String(g.submodules), nil
	case "first_parent":
		return starlark.Bool(g.firstParent), nil
	case "partial_fetch":
		return starlark.Bool(g.partialFetch), nil
	case "state":
		return starlark.String(g.state), nil
	case "review_state":
		return starlark.String(g.reviewState), nil
	case "primary_branch_migration":
		return starlark.Bool(g.primaryBranchMigration), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (g *GitHubOrigin) AttrNames() []string {
	return []string{
		"url",
		"ref",
		"branch",
		"submodules",
		"first_parent",
		"partial_fetch",
		"state",
		"review_state",
		"primary_branch_migration",
	}
}

// GitHubPrDestination represents a GitHub PR destination.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/git/GitHubPrDestination.java
type GitHubPrDestination struct {
	url                    string
	destinationRef         string
	prBranch               string
	title                  string
	body                   string
	draft                  bool
	primaryBranchMigration bool
	updateDescription      bool
	integrates             []*IntegrateChanges
}

// String implements starlark.Value.
func (g *GitHubPrDestination) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("url = %q", g.url))
	if g.destinationRef != "" {
		parts = append(parts, fmt.Sprintf("destination_ref = %q", g.destinationRef))
	}
	if g.prBranch != "" {
		parts = append(parts, fmt.Sprintf("pr_branch = %q", g.prBranch))
	}
	if g.title != "" {
		parts = append(parts, fmt.Sprintf("title = %q", g.title))
	}
	if g.body != "" {
		parts = append(parts, fmt.Sprintf("body = %q", g.body))
	}
	if g.draft {
		parts = append(parts, "draft = True")
	}
	return fmt.Sprintf("git.github_pr_destination(%s)", strings.Join(parts, ", "))
}

// Type implements starlark.Value.
func (g *GitHubPrDestination) Type() string {
	return "git.github_pr_destination"
}

// Freeze implements starlark.Value.
func (g *GitHubPrDestination) Freeze() {}

// Truth implements starlark.Value.
func (g *GitHubPrDestination) Truth() starlark.Bool {
	return starlark.True
}

// Hash implements starlark.Value.
func (g *GitHubPrDestination) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: git.github_pr_destination")
}

// URL returns the repository URL.
func (g *GitHubPrDestination) URL() string {
	return g.url
}

// DestinationRef returns the target branch (base branch for PR).
func (g *GitHubPrDestination) DestinationRef() string {
	return g.destinationRef
}

// PrBranch returns the PR branch name template.
func (g *GitHubPrDestination) PrBranch() string {
	return g.prBranch
}

// Title returns the PR title template.
func (g *GitHubPrDestination) Title() string {
	return g.title
}

// Body returns the PR body template.
func (g *GitHubPrDestination) Body() string {
	return g.body
}

// Draft returns whether to create draft PRs.
func (g *GitHubPrDestination) Draft() bool {
	return g.draft
}

// PrimaryBranchMigration returns whether primary branch migration mode is enabled.
func (g *GitHubPrDestination) PrimaryBranchMigration() bool {
	return g.primaryBranchMigration
}

// UpdateDescription returns whether to update PR description on subsequent pushes.
func (g *GitHubPrDestination) UpdateDescription() bool {
	return g.updateDescription
}

// Integrates returns the integrate changes configuration.
func (g *GitHubPrDestination) Integrates() []*IntegrateChanges {
	return g.integrates
}

// Attr implements starlark.HasAttrs.
func (g *GitHubPrDestination) Attr(name string) (starlark.Value, error) {
	switch name {
	case "url":
		return starlark.String(g.url), nil
	case "destination_ref":
		return starlark.String(g.destinationRef), nil
	case "pr_branch":
		return starlark.String(g.prBranch), nil
	case "title":
		return starlark.String(g.title), nil
	case "body":
		return starlark.String(g.body), nil
	case "draft":
		return starlark.Bool(g.draft), nil
	case "primary_branch_migration":
		return starlark.Bool(g.primaryBranchMigration), nil
	case "update_description":
		return starlark.Bool(g.updateDescription), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (g *GitHubPrDestination) AttrNames() []string {
	return []string{
		"url",
		"destination_ref",
		"pr_branch",
		"title",
		"body",
		"draft",
		"primary_branch_migration",
		"update_description",
	}
}

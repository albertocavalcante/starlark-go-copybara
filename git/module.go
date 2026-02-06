// Package git provides the git.* Starlark module for Copybara.
//
// The git module provides origins and destinations for Git repositories:
//   - git.origin() - Git repository origin
//   - git.destination() - Git repository destination
//   - git.github_origin() - GitHub origin
//   - git.github_pr_destination() - GitHub PR destination
//   - git.integrate() - Integration configuration
//
// Reference: https://github.com/google/copybara/tree/master/java/com/google/copybara/git
package git

import (
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// Module is the git.* Starlark module.
var Module = &starlarkstruct.Module{
	Name: "git",
	Members: starlark.StringDict{
		"origin":                starlark.NewBuiltin("git.origin", originFn),
		"destination":           starlark.NewBuiltin("git.destination", destinationFn),
		"github_origin":         starlark.NewBuiltin("git.github_origin", githubOriginFn),
		"github_pr_destination": starlark.NewBuiltin("git.github_pr_destination", githubPrDestinationFn),
		"integrate":             starlark.NewBuiltin("git.integrate", integrateFn),
	},
}

// originFn implements git.origin().
//
// Parameters:
//   - url (required): Repository URL
//   - ref (optional): Git reference (default: "master")
//   - submodules (optional): "YES", "NO", "RECURSIVE" (default: "NO")
//   - include_branch_commit_logs (optional): Include branch commit logs (default: false)
//   - first_parent (optional): Use first parent only (default: true)
//   - partial_fetch (optional): Enable partial fetch (default: false)
//   - primary_branch_migration (optional): Enable primary branch migration mode (default: false)
//
// Reference: GitOrigin.java
func originFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url                     string
		ref                     string = "master"
		submodules              string = "NO"
		includeBranchCommitLogs bool   = false
		firstParent             bool   = true
		partialFetch            bool   = false
		primaryBranchMigration  bool   = false
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"url", &url,
		"ref?", &ref,
		"submodules?", &submodules,
		"include_branch_commit_logs?", &includeBranchCommitLogs,
		"first_parent?", &firstParent,
		"partial_fetch?", &partialFetch,
		"primary_branch_migration?", &primaryBranchMigration,
	); err != nil {
		return nil, err
	}

	submoduleStrategy, err := ParseSubmoduleStrategy(submodules)
	if err != nil {
		return nil, err
	}

	return &Origin{
		url:                     url,
		ref:                     ref,
		submodules:              submoduleStrategy,
		includeBranchCommitLogs: includeBranchCommitLogs,
		firstParent:             firstParent,
		partialFetch:            partialFetch,
		primaryBranchMigration:  primaryBranchMigration,
	}, nil
}

// destinationFn implements git.destination().
//
// Parameters:
//   - url (required): Repository URL
//   - push (optional): Branch to push to
//   - fetch (optional): Branch to fetch from
//   - tag_name (optional): Tag name template
//   - tag_msg (optional): Tag message template
//   - skip_push (optional): Skip push (dry-run mode) (default: false)
//   - integrates (optional): List of git.integrate() configurations
//   - primary_branch_migration (optional): Enable primary branch migration mode (default: false)
//
// Reference: GitDestination.java
func destinationFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url                    string
		push                   string
		fetch                  string
		tagName                string
		tagMsg                 string
		skipPush               bool = false
		branch                 string
		integratesValue        *starlark.List
		primaryBranchMigration bool = false
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"url", &url,
		"push?", &push,
		"fetch?", &fetch,
		"tag_name?", &tagName,
		"tag_msg?", &tagMsg,
		"skip_push?", &skipPush,
		"branch?", &branch,
		"integrates?", &integratesValue,
		"primary_branch_migration?", &primaryBranchMigration,
	); err != nil {
		return nil, err
	}

	// Default branch to "main" if not specified
	if push == "" && branch != "" {
		push = branch
	}
	if fetch == "" && branch != "" {
		fetch = branch
	}

	// Parse integrates list
	var integrates []*IntegrateChanges
	if integratesValue != nil {
		for i := 0; i < integratesValue.Len(); i++ {
			item := integratesValue.Index(i)
			ic, ok := item.(*IntegrateChanges)
			if !ok {
				return nil, fmt.Errorf("integrates[%d]: expected git.integrate, got %s", i, item.Type())
			}
			integrates = append(integrates, ic)
		}
	}

	return &Destination{
		url:                    url,
		push:                   push,
		fetch:                  fetch,
		tagName:                tagName,
		tagMsg:                 tagMsg,
		skipPush:               skipPush,
		primaryBranchMigration: primaryBranchMigration,
		integrates:             integrates,
	}, nil
}

// githubOriginFn implements git.github_origin().
//
// Parameters:
//   - url (required): GitHub repository URL
//   - ref (optional): Git reference (default: "master")
//   - branch (optional): Alternative to ref - the branch to track
//   - submodules (optional): "YES", "NO", "RECURSIVE" (default: "NO")
//   - first_parent (optional): Use first parent only (default: true)
//   - partial_fetch (optional): Enable partial fetch (default: false)
//   - state (optional): PR state filter: "OPEN", "CLOSED", "ALL" (default: "OPEN")
//   - review_state (optional): Required review state (default: "ANY")
//   - primary_branch_migration (optional): Enable primary branch migration mode (default: false)
//
// Reference: GitHubPrOrigin.java
func githubOriginFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url                    string
		ref                    string = "master"
		branch                 string
		submodules             string = "NO"
		firstParent            bool   = true
		partialFetch           bool   = false
		state                  string = "OPEN"
		reviewState            string = "ANY"
		primaryBranchMigration bool   = false
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"url", &url,
		"ref?", &ref,
		"branch?", &branch,
		"submodules?", &submodules,
		"first_parent?", &firstParent,
		"partial_fetch?", &partialFetch,
		"state?", &state,
		"review_state?", &reviewState,
		"primary_branch_migration?", &primaryBranchMigration,
	); err != nil {
		return nil, err
	}

	// Branch overrides ref if provided
	if branch != "" {
		ref = branch
	}

	submoduleStrategy, err := ParseSubmoduleStrategy(submodules)
	if err != nil {
		return nil, err
	}

	stateFilter, err := ParseStateFilter(state)
	if err != nil {
		return nil, err
	}

	reviewStateVal, err := ParseReviewState(reviewState)
	if err != nil {
		return nil, err
	}

	return &GitHubOrigin{
		url:                    url,
		ref:                    ref,
		branch:                 branch,
		submodules:             submoduleStrategy,
		firstParent:            firstParent,
		partialFetch:           partialFetch,
		state:                  stateFilter,
		reviewState:            reviewStateVal,
		primaryBranchMigration: primaryBranchMigration,
	}, nil
}

// githubPrDestinationFn implements git.github_pr_destination().
//
// Parameters:
//   - url (required): GitHub repository URL
//   - destination_ref (optional): Target branch for PR (default: "main")
//   - pr_branch (optional): PR branch name template
//   - title (optional): PR title template
//   - body (optional): PR body template
//   - draft (optional): Create draft PRs (default: false)
//   - integrates (optional): List of git.integrate() configurations
//   - primary_branch_migration (optional): Enable primary branch migration mode (default: false)
//   - update_description (optional): Update PR description on subsequent pushes (default: true)
//
// Reference: GitHubPrDestination.java
func githubPrDestinationFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		url                    string
		destinationRef         string = "main"
		prBranch               string
		title                  string
		body                   string
		draft                  bool = false
		integratesValue        *starlark.List
		primaryBranchMigration bool = false
		updateDescription      bool = true
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"url", &url,
		"destination_ref?", &destinationRef,
		"pr_branch?", &prBranch,
		"title?", &title,
		"body?", &body,
		"draft?", &draft,
		"integrates?", &integratesValue,
		"primary_branch_migration?", &primaryBranchMigration,
		"update_description?", &updateDescription,
	); err != nil {
		return nil, err
	}

	// Parse integrates list
	var integrates []*IntegrateChanges
	if integratesValue != nil {
		for i := 0; i < integratesValue.Len(); i++ {
			item := integratesValue.Index(i)
			ic, ok := item.(*IntegrateChanges)
			if !ok {
				return nil, fmt.Errorf("integrates[%d]: expected git.integrate, got %s", i, item.Type())
			}
			integrates = append(integrates, ic)
		}
	}

	return &GitHubPrDestination{
		url:                    url,
		destinationRef:         destinationRef,
		prBranch:               prBranch,
		title:                  title,
		body:                   body,
		draft:                  draft,
		primaryBranchMigration: primaryBranchMigration,
		updateDescription:      updateDescription,
		integrates:             integrates,
	}, nil
}

// integrateFn implements git.integrate().
//
// Parameters:
//   - label (optional): The label containing the URL to integrate (default: "COPYBARA_INTEGRATE_REVIEW")
//   - strategy (optional): Integration strategy (default: "FAKE_MERGE_AND_INCLUDE_FILES")
//     - "FAKE_MERGE": Add as parent but ignore files
//     - "FAKE_MERGE_AND_INCLUDE_FILES": Fake merge but include non-destination files
//     - "INCLUDE_FILES": Include non-destination files without merge commit
//   - ignore_errors (optional): Ignore integration errors (default: true)
//
// Reference: GitIntegrateChanges.java
func integrateFn(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		label        string = DefaultIntegrateLabel
		strategy     string = "FAKE_MERGE_AND_INCLUDE_FILES"
		ignoreErrors bool   = true
	)

	if err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"label?", &label,
		"strategy?", &strategy,
		"ignore_errors?", &ignoreErrors,
	); err != nil {
		return nil, err
	}

	strategyVal, err := ParseIntegrateStrategy(strategy)
	if err != nil {
		return nil, err
	}

	return NewIntegrateChanges(label, strategyVal, ignoreErrors), nil
}

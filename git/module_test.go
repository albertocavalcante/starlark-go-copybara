package git

import (
	"testing"

	"go.starlark.net/starlark"
)

func TestOriginFn(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantURL        string
		wantRef        string
		wantSubmodules SubmoduleStrategy
		wantFirstParent bool
		wantPartialFetch bool
		wantErr        bool
	}{
		{
			name:            "minimal",
			code:            `git.origin(url = "https://github.com/example/repo")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantSubmodules:  SubmoduleNo,
			wantFirstParent: true,
			wantPartialFetch: false,
		},
		{
			name:            "with ref",
			code:            `git.origin(url = "https://github.com/example/repo", ref = "main")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "main",
			wantSubmodules:  SubmoduleNo,
			wantFirstParent: true,
		},
		{
			name:            "with submodules YES",
			code:            `git.origin(url = "https://github.com/example/repo", submodules = "YES")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantSubmodules:  SubmoduleYes,
			wantFirstParent: true,
		},
		{
			name:            "with submodules RECURSIVE",
			code:            `git.origin(url = "https://github.com/example/repo", submodules = "RECURSIVE")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantSubmodules:  SubmoduleRecursive,
			wantFirstParent: true,
		},
		{
			name:            "with first_parent false",
			code:            `git.origin(url = "https://github.com/example/repo", first_parent = False)`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantSubmodules:  SubmoduleNo,
			wantFirstParent: false,
		},
		{
			name:             "with partial_fetch",
			code:             `git.origin(url = "https://github.com/example/repo", partial_fetch = True)`,
			wantURL:          "https://github.com/example/repo",
			wantRef:          "master",
			wantSubmodules:   SubmoduleNo,
			wantFirstParent:  true,
			wantPartialFetch: true,
		},
		{
			name:    "missing url",
			code:    `git.origin()`,
			wantErr: true,
		},
		{
			name:    "invalid submodules",
			code:    `git.origin(url = "https://github.com/example/repo", submodules = "INVALID")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalGitExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			origin, ok := result.(*Origin)
			if !ok {
				t.Fatalf("expected *Origin, got %T", result)
			}

			if origin.URL() != tt.wantURL {
				t.Errorf("URL() = %q, want %q", origin.URL(), tt.wantURL)
			}
			if origin.Ref() != tt.wantRef {
				t.Errorf("Ref() = %q, want %q", origin.Ref(), tt.wantRef)
			}
			if origin.Submodules() != tt.wantSubmodules {
				t.Errorf("Submodules() = %q, want %q", origin.Submodules(), tt.wantSubmodules)
			}
			if origin.FirstParent() != tt.wantFirstParent {
				t.Errorf("FirstParent() = %v, want %v", origin.FirstParent(), tt.wantFirstParent)
			}
			if origin.PartialFetch() != tt.wantPartialFetch {
				t.Errorf("PartialFetch() = %v, want %v", origin.PartialFetch(), tt.wantPartialFetch)
			}
		})
	}
}

func TestDestinationFn(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		wantURL     string
		wantPush    string
		wantFetch   string
		wantTagName string
		wantTagMsg  string
		wantSkipPush bool
		wantErr     bool
	}{
		{
			name:    "minimal",
			code:    `git.destination(url = "https://github.com/example/repo")`,
			wantURL: "https://github.com/example/repo",
		},
		{
			name:      "with push and fetch",
			code:      `git.destination(url = "https://github.com/example/repo", push = "main", fetch = "main")`,
			wantURL:   "https://github.com/example/repo",
			wantPush:  "main",
			wantFetch: "main",
		},
		{
			name:      "with branch shorthand",
			code:      `git.destination(url = "https://github.com/example/repo", branch = "develop")`,
			wantURL:   "https://github.com/example/repo",
			wantPush:  "develop",
			wantFetch: "develop",
		},
		{
			name:        "with tag",
			code:        `git.destination(url = "https://github.com/example/repo", tag_name = "v${VERSION}", tag_msg = "Release ${VERSION}")`,
			wantURL:     "https://github.com/example/repo",
			wantTagName: "v${VERSION}",
			wantTagMsg:  "Release ${VERSION}",
		},
		{
			name:         "with skip_push",
			code:         `git.destination(url = "https://github.com/example/repo", skip_push = True)`,
			wantURL:      "https://github.com/example/repo",
			wantSkipPush: true,
		},
		{
			name:    "missing url",
			code:    `git.destination()`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalGitExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			dest, ok := result.(*Destination)
			if !ok {
				t.Fatalf("expected *Destination, got %T", result)
			}

			if dest.URL() != tt.wantURL {
				t.Errorf("URL() = %q, want %q", dest.URL(), tt.wantURL)
			}
			if dest.Push() != tt.wantPush {
				t.Errorf("Push() = %q, want %q", dest.Push(), tt.wantPush)
			}
			if dest.Fetch() != tt.wantFetch {
				t.Errorf("Fetch() = %q, want %q", dest.Fetch(), tt.wantFetch)
			}
			if dest.TagName() != tt.wantTagName {
				t.Errorf("TagName() = %q, want %q", dest.TagName(), tt.wantTagName)
			}
			if dest.TagMsg() != tt.wantTagMsg {
				t.Errorf("TagMsg() = %q, want %q", dest.TagMsg(), tt.wantTagMsg)
			}
			if dest.SkipPush() != tt.wantSkipPush {
				t.Errorf("SkipPush() = %v, want %v", dest.SkipPush(), tt.wantSkipPush)
			}
		})
	}
}

func TestGitHubOriginFn(t *testing.T) {
	tests := []struct {
		name            string
		code            string
		wantURL         string
		wantRef         string
		wantBranch      string
		wantState       StateFilter
		wantReviewState ReviewState
		wantErr         bool
	}{
		{
			name:            "minimal",
			code:            `git.github_origin(url = "https://github.com/example/repo")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantState:       StateFilterOpen,
			wantReviewState: ReviewStateAny,
		},
		{
			name:            "with ref",
			code:            `git.github_origin(url = "https://github.com/example/repo", ref = "main")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "main",
			wantState:       StateFilterOpen,
			wantReviewState: ReviewStateAny,
		},
		{
			name:            "with branch",
			code:            `git.github_origin(url = "https://github.com/example/repo", branch = "develop")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "develop",
			wantBranch:      "develop",
			wantState:       StateFilterOpen,
			wantReviewState: ReviewStateAny,
		},
		{
			name:            "with state CLOSED",
			code:            `git.github_origin(url = "https://github.com/example/repo", state = "CLOSED")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantState:       StateFilterClosed,
			wantReviewState: ReviewStateAny,
		},
		{
			name:            "with state ALL",
			code:            `git.github_origin(url = "https://github.com/example/repo", state = "ALL")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantState:       StateFilterAll,
			wantReviewState: ReviewStateAny,
		},
		{
			name:            "with review_state HEAD_COMMIT_APPROVED",
			code:            `git.github_origin(url = "https://github.com/example/repo", review_state = "HEAD_COMMIT_APPROVED")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantState:       StateFilterOpen,
			wantReviewState: ReviewStateHeadCommitApproved,
		},
		{
			name:            "with review_state ANY_COMMIT_APPROVED",
			code:            `git.github_origin(url = "https://github.com/example/repo", review_state = "ANY_COMMIT_APPROVED")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantState:       StateFilterOpen,
			wantReviewState: ReviewStateAnyCommitApproved,
		},
		{
			name:            "with review_state HAS_REVIEWERS",
			code:            `git.github_origin(url = "https://github.com/example/repo", review_state = "HAS_REVIEWERS")`,
			wantURL:         "https://github.com/example/repo",
			wantRef:         "master",
			wantState:       StateFilterOpen,
			wantReviewState: ReviewStateHasReviewers,
		},
		{
			name:    "missing url",
			code:    `git.github_origin()`,
			wantErr: true,
		},
		{
			name:    "invalid state",
			code:    `git.github_origin(url = "https://github.com/example/repo", state = "INVALID")`,
			wantErr: true,
		},
		{
			name:    "invalid review_state",
			code:    `git.github_origin(url = "https://github.com/example/repo", review_state = "INVALID")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalGitExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			origin, ok := result.(*GitHubOrigin)
			if !ok {
				t.Fatalf("expected *GitHubOrigin, got %T", result)
			}

			if origin.URL() != tt.wantURL {
				t.Errorf("URL() = %q, want %q", origin.URL(), tt.wantURL)
			}
			if origin.Ref() != tt.wantRef {
				t.Errorf("Ref() = %q, want %q", origin.Ref(), tt.wantRef)
			}
			if origin.Branch() != tt.wantBranch {
				t.Errorf("Branch() = %q, want %q", origin.Branch(), tt.wantBranch)
			}
			if origin.State() != tt.wantState {
				t.Errorf("State() = %q, want %q", origin.State(), tt.wantState)
			}
			if origin.ReviewState() != tt.wantReviewState {
				t.Errorf("ReviewState() = %q, want %q", origin.ReviewState(), tt.wantReviewState)
			}
		})
	}
}

func TestGitHubPrDestinationFn(t *testing.T) {
	tests := []struct {
		name               string
		code               string
		wantURL            string
		wantDestinationRef string
		wantPrBranch       string
		wantTitle          string
		wantBody           string
		wantDraft          bool
		wantUpdateDesc     bool
		wantErr            bool
	}{
		{
			name:               "minimal",
			code:               `git.github_pr_destination(url = "https://github.com/example/repo")`,
			wantURL:            "https://github.com/example/repo",
			wantDestinationRef: "main",
			wantUpdateDesc:     true,
		},
		{
			name:               "with destination_ref",
			code:               `git.github_pr_destination(url = "https://github.com/example/repo", destination_ref = "develop")`,
			wantURL:            "https://github.com/example/repo",
			wantDestinationRef: "develop",
			wantUpdateDesc:     true,
		},
		{
			name:               "with pr_branch",
			code:               `git.github_pr_destination(url = "https://github.com/example/repo", pr_branch = "copybara/${CONTEXT_REFERENCE}")`,
			wantURL:            "https://github.com/example/repo",
			wantDestinationRef: "main",
			wantPrBranch:       "copybara/${CONTEXT_REFERENCE}",
			wantUpdateDesc:     true,
		},
		{
			name:               "with title and body",
			code:               `git.github_pr_destination(url = "https://github.com/example/repo", title = "Migrate ${LABEL}", body = "Automated migration")`,
			wantURL:            "https://github.com/example/repo",
			wantDestinationRef: "main",
			wantTitle:          "Migrate ${LABEL}",
			wantBody:           "Automated migration",
			wantUpdateDesc:     true,
		},
		{
			name:               "with draft",
			code:               `git.github_pr_destination(url = "https://github.com/example/repo", draft = True)`,
			wantURL:            "https://github.com/example/repo",
			wantDestinationRef: "main",
			wantDraft:          true,
			wantUpdateDesc:     true,
		},
		{
			name:               "with update_description false",
			code:               `git.github_pr_destination(url = "https://github.com/example/repo", update_description = False)`,
			wantURL:            "https://github.com/example/repo",
			wantDestinationRef: "main",
			wantUpdateDesc:     false,
		},
		{
			name:    "missing url",
			code:    `git.github_pr_destination()`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalGitExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			dest, ok := result.(*GitHubPrDestination)
			if !ok {
				t.Fatalf("expected *GitHubPrDestination, got %T", result)
			}

			if dest.URL() != tt.wantURL {
				t.Errorf("URL() = %q, want %q", dest.URL(), tt.wantURL)
			}
			if dest.DestinationRef() != tt.wantDestinationRef {
				t.Errorf("DestinationRef() = %q, want %q", dest.DestinationRef(), tt.wantDestinationRef)
			}
			if dest.PrBranch() != tt.wantPrBranch {
				t.Errorf("PrBranch() = %q, want %q", dest.PrBranch(), tt.wantPrBranch)
			}
			if dest.Title() != tt.wantTitle {
				t.Errorf("Title() = %q, want %q", dest.Title(), tt.wantTitle)
			}
			if dest.Body() != tt.wantBody {
				t.Errorf("Body() = %q, want %q", dest.Body(), tt.wantBody)
			}
			if dest.Draft() != tt.wantDraft {
				t.Errorf("Draft() = %v, want %v", dest.Draft(), tt.wantDraft)
			}
			if dest.UpdateDescription() != tt.wantUpdateDesc {
				t.Errorf("UpdateDescription() = %v, want %v", dest.UpdateDescription(), tt.wantUpdateDesc)
			}
		})
	}
}

func TestIntegrateFn(t *testing.T) {
	tests := []struct {
		name             string
		code             string
		wantLabel        string
		wantStrategy     IntegrateStrategy
		wantIgnoreErrors bool
		wantErr          bool
	}{
		{
			name:             "defaults",
			code:             `git.integrate()`,
			wantLabel:        DefaultIntegrateLabel,
			wantStrategy:     StrategyFakeMergeAndIncludeFiles,
			wantIgnoreErrors: true,
		},
		{
			name:             "with label",
			code:             `git.integrate(label = "CUSTOM_LABEL")`,
			wantLabel:        "CUSTOM_LABEL",
			wantStrategy:     StrategyFakeMergeAndIncludeFiles,
			wantIgnoreErrors: true,
		},
		{
			name:             "with FAKE_MERGE strategy",
			code:             `git.integrate(strategy = "FAKE_MERGE")`,
			wantLabel:        DefaultIntegrateLabel,
			wantStrategy:     StrategyFakeMerge,
			wantIgnoreErrors: true,
		},
		{
			name:             "with INCLUDE_FILES strategy",
			code:             `git.integrate(strategy = "INCLUDE_FILES")`,
			wantLabel:        DefaultIntegrateLabel,
			wantStrategy:     StrategyIncludeFiles,
			wantIgnoreErrors: true,
		},
		{
			name:             "with ignore_errors false",
			code:             `git.integrate(ignore_errors = False)`,
			wantLabel:        DefaultIntegrateLabel,
			wantStrategy:     StrategyFakeMergeAndIncludeFiles,
			wantIgnoreErrors: false,
		},
		{
			name:             "with all params",
			code:             `git.integrate(label = "MY_LABEL", strategy = "FAKE_MERGE", ignore_errors = False)`,
			wantLabel:        "MY_LABEL",
			wantStrategy:     StrategyFakeMerge,
			wantIgnoreErrors: false,
		},
		{
			name:    "invalid strategy",
			code:    `git.integrate(strategy = "INVALID")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalGitExpr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			integrate, ok := result.(*IntegrateChanges)
			if !ok {
				t.Fatalf("expected *IntegrateChanges, got %T", result)
			}

			if integrate.Label() != tt.wantLabel {
				t.Errorf("Label() = %q, want %q", integrate.Label(), tt.wantLabel)
			}
			if integrate.Strategy() != tt.wantStrategy {
				t.Errorf("Strategy() = %q, want %q", integrate.Strategy(), tt.wantStrategy)
			}
			if integrate.IgnoreErrors() != tt.wantIgnoreErrors {
				t.Errorf("IgnoreErrors() = %v, want %v", integrate.IgnoreErrors(), tt.wantIgnoreErrors)
			}
		})
	}
}

func TestDestinationWithIntegrates(t *testing.T) {
	code := `git.destination(
		url = "https://github.com/example/repo",
		integrates = [
			git.integrate(),
			git.integrate(label = "CUSTOM", strategy = "FAKE_MERGE"),
		],
	)`

	result, err := evalGitExpr(t, code)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	dest, ok := result.(*Destination)
	if !ok {
		t.Fatalf("expected *Destination, got %T", result)
	}

	integrates := dest.Integrates()
	if len(integrates) != 2 {
		t.Fatalf("expected 2 integrates, got %d", len(integrates))
	}

	if integrates[0].Label() != DefaultIntegrateLabel {
		t.Errorf("integrates[0].Label() = %q, want %q", integrates[0].Label(), DefaultIntegrateLabel)
	}
	if integrates[1].Label() != "CUSTOM" {
		t.Errorf("integrates[1].Label() = %q, want %q", integrates[1].Label(), "CUSTOM")
	}
	if integrates[1].Strategy() != StrategyFakeMerge {
		t.Errorf("integrates[1].Strategy() = %q, want %q", integrates[1].Strategy(), StrategyFakeMerge)
	}
}

func TestGitHubPrDestinationWithIntegrates(t *testing.T) {
	code := `git.github_pr_destination(
		url = "https://github.com/example/repo",
		integrates = [git.integrate()],
	)`

	result, err := evalGitExpr(t, code)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	dest, ok := result.(*GitHubPrDestination)
	if !ok {
		t.Fatalf("expected *GitHubPrDestination, got %T", result)
	}

	integrates := dest.Integrates()
	if len(integrates) != 1 {
		t.Fatalf("expected 1 integrate, got %d", len(integrates))
	}
}

func TestOriginString(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "minimal",
			code: `git.origin(url = "https://github.com/example/repo")`,
			want: `git.origin(url = "https://github.com/example/repo", ref = "master")`,
		},
		{
			name: "with submodules",
			code: `git.origin(url = "https://github.com/example/repo", submodules = "YES")`,
			want: `git.origin(url = "https://github.com/example/repo", ref = "master", submodules = "YES")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if err != nil {
				t.Fatalf("evalGitExpr() error = %v", err)
			}
			if result.String() != tt.want {
				t.Errorf("String() = %q, want %q", result.String(), tt.want)
			}
		})
	}
}

func TestOriginType(t *testing.T) {
	tests := []struct {
		code     string
		wantType string
	}{
		{`git.origin(url = "https://github.com/example/repo")`, "git.origin"},
		{`git.destination(url = "https://github.com/example/repo")`, "git.destination"},
		{`git.github_origin(url = "https://github.com/example/repo")`, "git.github_origin"},
		{`git.github_pr_destination(url = "https://github.com/example/repo")`, "git.github_pr_destination"},
		{`git.integrate()`, "git.integrate"},
	}

	for _, tt := range tests {
		t.Run(tt.wantType, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if err != nil {
				t.Fatalf("evalGitExpr() error = %v", err)
			}
			if result.Type() != tt.wantType {
				t.Errorf("Type() = %q, want %q", result.Type(), tt.wantType)
			}
		})
	}
}

func TestOriginAttr(t *testing.T) {
	result, err := evalGitExpr(t, `git.origin(url = "https://github.com/example/repo", ref = "main")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	origin := result.(*Origin)

	urlVal, err := origin.Attr("url")
	if err != nil {
		t.Errorf("Attr(url) error = %v", err)
	}
	if urlVal.(starlark.String).GoString() != "https://github.com/example/repo" {
		t.Errorf("Attr(url) = %q, want %q", urlVal, "https://github.com/example/repo")
	}

	refVal, err := origin.Attr("ref")
	if err != nil {
		t.Errorf("Attr(ref) error = %v", err)
	}
	if refVal.(starlark.String).GoString() != "main" {
		t.Errorf("Attr(ref) = %q, want %q", refVal, "main")
	}

	unknownVal, err := origin.Attr("unknown")
	if err != nil {
		t.Errorf("Attr(unknown) error = %v", err)
	}
	if unknownVal != nil {
		t.Errorf("Attr(unknown) = %v, want nil", unknownVal)
	}
}

func TestGitHubOriginAttr(t *testing.T) {
	result, err := evalGitExpr(t, `git.github_origin(url = "https://github.com/example/repo", review_state = "HEAD_COMMIT_APPROVED")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	origin := result.(*GitHubOrigin)

	reviewStateVal, err := origin.Attr("review_state")
	if err != nil {
		t.Errorf("Attr(review_state) error = %v", err)
	}
	if reviewStateVal.(starlark.String).GoString() != string(ReviewStateHeadCommitApproved) {
		t.Errorf("Attr(review_state) = %q, want %q", reviewStateVal, ReviewStateHeadCommitApproved)
	}
}

func TestParseSubmoduleStrategy(t *testing.T) {
	tests := []struct {
		input   string
		want    SubmoduleStrategy
		wantErr bool
	}{
		{"NO", SubmoduleNo, false},
		{"no", SubmoduleNo, false},
		{"YES", SubmoduleYes, false},
		{"yes", SubmoduleYes, false},
		{"RECURSIVE", SubmoduleRecursive, false},
		{"recursive", SubmoduleRecursive, false},
		{"", SubmoduleNo, false},
		{"INVALID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSubmoduleStrategy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSubmoduleStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseSubmoduleStrategy() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseIntegrateStrategy(t *testing.T) {
	tests := []struct {
		input   string
		want    IntegrateStrategy
		wantErr bool
	}{
		{"FAKE_MERGE", StrategyFakeMerge, false},
		{"fake_merge", StrategyFakeMerge, false},
		{"FAKE_MERGE_AND_INCLUDE_FILES", StrategyFakeMergeAndIncludeFiles, false},
		{"INCLUDE_FILES", StrategyIncludeFiles, false},
		{"", StrategyFakeMergeAndIncludeFiles, false},
		{"INVALID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseIntegrateStrategy(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIntegrateStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseIntegrateStrategy() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseReviewState(t *testing.T) {
	tests := []struct {
		input   string
		want    ReviewState
		wantErr bool
	}{
		{"HEAD_COMMIT_APPROVED", ReviewStateHeadCommitApproved, false},
		{"ANY_COMMIT_APPROVED", ReviewStateAnyCommitApproved, false},
		{"HAS_REVIEWERS", ReviewStateHasReviewers, false},
		{"ANY", ReviewStateAny, false},
		{"", ReviewStateAny, false},
		{"INVALID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseReviewState(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReviewState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseReviewState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStateFilter(t *testing.T) {
	tests := []struct {
		input   string
		want    StateFilter
		wantErr bool
	}{
		{"OPEN", StateFilterOpen, false},
		{"CLOSED", StateFilterClosed, false},
		{"ALL", StateFilterAll, false},
		{"", StateFilterOpen, false},
		{"INVALID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseStateFilter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStateFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseStateFilter() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOriginTruthAndHash(t *testing.T) {
	result, err := evalGitExpr(t, `git.origin(url = "https://github.com/example/repo")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	origin := result.(*Origin)

	// Test Truth
	if origin.Truth() != starlark.True {
		t.Errorf("Truth() = %v, want True", origin.Truth())
	}

	// Test Hash (should fail)
	_, err = origin.Hash()
	if err == nil {
		t.Error("Hash() should return an error")
	}

	// Test Freeze (should not panic)
	origin.Freeze()
}

func TestDestinationTruthAndHash(t *testing.T) {
	result, err := evalGitExpr(t, `git.destination(url = "https://github.com/example/repo")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	dest := result.(*Destination)

	if dest.Truth() != starlark.True {
		t.Errorf("Truth() = %v, want True", dest.Truth())
	}

	_, err = dest.Hash()
	if err == nil {
		t.Error("Hash() should return an error")
	}

	dest.Freeze()
}

func TestGitHubOriginTruthAndHash(t *testing.T) {
	result, err := evalGitExpr(t, `git.github_origin(url = "https://github.com/example/repo")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	origin := result.(*GitHubOrigin)

	if origin.Truth() != starlark.True {
		t.Errorf("Truth() = %v, want True", origin.Truth())
	}

	_, err = origin.Hash()
	if err == nil {
		t.Error("Hash() should return an error")
	}

	origin.Freeze()
}

func TestGitHubPrDestinationTruthAndHash(t *testing.T) {
	result, err := evalGitExpr(t, `git.github_pr_destination(url = "https://github.com/example/repo")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	dest := result.(*GitHubPrDestination)

	if dest.Truth() != starlark.True {
		t.Errorf("Truth() = %v, want True", dest.Truth())
	}

	_, err = dest.Hash()
	if err == nil {
		t.Error("Hash() should return an error")
	}

	dest.Freeze()
}

func TestIntegrateTruthAndHash(t *testing.T) {
	result, err := evalGitExpr(t, `git.integrate()`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	integrate := result.(*IntegrateChanges)

	if integrate.Truth() != starlark.True {
		t.Errorf("Truth() = %v, want True", integrate.Truth())
	}

	_, err = integrate.Hash()
	if err == nil {
		t.Error("Hash() should return an error")
	}

	integrate.Freeze()
}

func TestIntegrateString(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "defaults",
			code: `git.integrate()`,
			want: `git.integrate()`,
		},
		{
			name: "with custom label",
			code: `git.integrate(label = "CUSTOM_LABEL")`,
			want: `git.integrate(label = "CUSTOM_LABEL")`,
		},
		{
			name: "with FAKE_MERGE",
			code: `git.integrate(strategy = "FAKE_MERGE")`,
			want: `git.integrate(strategy = "FAKE_MERGE")`,
		},
		{
			name: "with ignore_errors false",
			code: `git.integrate(ignore_errors = False)`,
			want: `git.integrate(ignore_errors = False)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if err != nil {
				t.Fatalf("evalGitExpr() error = %v", err)
			}
			if result.String() != tt.want {
				t.Errorf("String() = %q, want %q", result.String(), tt.want)
			}
		})
	}
}

func TestDestinationString(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "minimal",
			code: `git.destination(url = "https://github.com/example/repo")`,
			want: `git.destination(url = "https://github.com/example/repo")`,
		},
		{
			name: "with push and fetch",
			code: `git.destination(url = "https://github.com/example/repo", push = "main", fetch = "main")`,
			want: `git.destination(url = "https://github.com/example/repo", push = "main", fetch = "main")`,
		},
		{
			name: "with skip_push",
			code: `git.destination(url = "https://github.com/example/repo", skip_push = True)`,
			want: `git.destination(url = "https://github.com/example/repo", skip_push = True)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if err != nil {
				t.Fatalf("evalGitExpr() error = %v", err)
			}
			if result.String() != tt.want {
				t.Errorf("String() = %q, want %q", result.String(), tt.want)
			}
		})
	}
}

func TestGitHubOriginString(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "minimal",
			code: `git.github_origin(url = "https://github.com/example/repo")`,
			want: `git.github_origin(url = "https://github.com/example/repo", ref = "master")`,
		},
		{
			name: "with review_state",
			code: `git.github_origin(url = "https://github.com/example/repo", review_state = "HEAD_COMMIT_APPROVED")`,
			want: `git.github_origin(url = "https://github.com/example/repo", ref = "master", review_state = "HEAD_COMMIT_APPROVED")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if err != nil {
				t.Fatalf("evalGitExpr() error = %v", err)
			}
			if result.String() != tt.want {
				t.Errorf("String() = %q, want %q", result.String(), tt.want)
			}
		})
	}
}

func TestGitHubPrDestinationString(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{
			name: "minimal",
			code: `git.github_pr_destination(url = "https://github.com/example/repo")`,
			want: `git.github_pr_destination(url = "https://github.com/example/repo", destination_ref = "main")`,
		},
		{
			name: "with draft",
			code: `git.github_pr_destination(url = "https://github.com/example/repo", draft = True)`,
			want: `git.github_pr_destination(url = "https://github.com/example/repo", destination_ref = "main", draft = True)`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evalGitExpr(t, tt.code)
			if err != nil {
				t.Fatalf("evalGitExpr() error = %v", err)
			}
			if result.String() != tt.want {
				t.Errorf("String() = %q, want %q", result.String(), tt.want)
			}
		})
	}
}

func TestDestinationAttr(t *testing.T) {
	result, err := evalGitExpr(t, `git.destination(url = "https://github.com/example/repo", push = "main", skip_push = True)`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	dest := result.(*Destination)

	tests := []struct {
		attr string
		want starlark.Value
	}{
		{"url", starlark.String("https://github.com/example/repo")},
		{"push", starlark.String("main")},
		{"skip_push", starlark.Bool(true)},
	}

	for _, tt := range tests {
		val, err := dest.Attr(tt.attr)
		if err != nil {
			t.Errorf("Attr(%q) error = %v", tt.attr, err)
		}
		if val != tt.want {
			t.Errorf("Attr(%q) = %v, want %v", tt.attr, val, tt.want)
		}
	}

	// Test AttrNames
	names := dest.AttrNames()
	if len(names) == 0 {
		t.Error("AttrNames() returned empty slice")
	}
}

func TestIntegrateAttr(t *testing.T) {
	result, err := evalGitExpr(t, `git.integrate(label = "MY_LABEL", strategy = "FAKE_MERGE")`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	integrate := result.(*IntegrateChanges)

	tests := []struct {
		attr string
		want starlark.Value
	}{
		{"label", starlark.String("MY_LABEL")},
		{"strategy", starlark.String(StrategyFakeMerge)},
		{"ignore_errors", starlark.Bool(true)},
	}

	for _, tt := range tests {
		val, err := integrate.Attr(tt.attr)
		if err != nil {
			t.Errorf("Attr(%q) error = %v", tt.attr, err)
		}
		if val != tt.want {
			t.Errorf("Attr(%q) = %v, want %v", tt.attr, val, tt.want)
		}
	}

	// Test unknown attr
	val, err := integrate.Attr("unknown")
	if err != nil || val != nil {
		t.Errorf("Attr(unknown) = %v, %v; want nil, nil", val, err)
	}

	// Test AttrNames
	names := integrate.AttrNames()
	if len(names) != 3 {
		t.Errorf("AttrNames() returned %d items, want 3", len(names))
	}
}

func TestGitHubPrDestinationAttr(t *testing.T) {
	result, err := evalGitExpr(t, `git.github_pr_destination(url = "https://github.com/example/repo", title = "My PR", draft = True)`)
	if err != nil {
		t.Fatalf("evalGitExpr() error = %v", err)
	}

	dest := result.(*GitHubPrDestination)

	tests := []struct {
		attr string
		want starlark.Value
	}{
		{"url", starlark.String("https://github.com/example/repo")},
		{"title", starlark.String("My PR")},
		{"draft", starlark.Bool(true)},
		{"update_description", starlark.Bool(true)},
	}

	for _, tt := range tests {
		val, err := dest.Attr(tt.attr)
		if err != nil {
			t.Errorf("Attr(%q) error = %v", tt.attr, err)
		}
		if val != tt.want {
			t.Errorf("Attr(%q) = %v, want %v", tt.attr, val, tt.want)
		}
	}
}

func TestNewIntegrateChanges(t *testing.T) {
	// Test with empty label (should use default)
	ic := NewIntegrateChanges("", StrategyFakeMerge, true)
	if ic.Label() != DefaultIntegrateLabel {
		t.Errorf("Label() = %q, want %q", ic.Label(), DefaultIntegrateLabel)
	}

	// Test with custom label
	ic = NewIntegrateChanges("CUSTOM", StrategyIncludeFiles, false)
	if ic.Label() != "CUSTOM" {
		t.Errorf("Label() = %q, want %q", ic.Label(), "CUSTOM")
	}
	if ic.Strategy() != StrategyIncludeFiles {
		t.Errorf("Strategy() = %q, want %q", ic.Strategy(), StrategyIncludeFiles)
	}
	if ic.IgnoreErrors() != false {
		t.Errorf("IgnoreErrors() = %v, want %v", ic.IgnoreErrors(), false)
	}
}

func TestModuleMembers(t *testing.T) {
	// Verify all expected members are present
	expectedMembers := []string{
		"origin",
		"destination",
		"github_origin",
		"github_pr_destination",
		"integrate",
	}

	for _, name := range expectedMembers {
		if _, ok := Module.Members[name]; !ok {
			t.Errorf("Module.Members[%q] not found", name)
		}
	}

	// Verify module name
	if Module.Name != "git" {
		t.Errorf("Module.Name = %q, want %q", Module.Name, "git")
	}
}

// evalGitExpr evaluates a Starlark expression with the git module available.
func evalGitExpr(t *testing.T, expr string) (starlark.Value, error) {
	t.Helper()

	predeclared := starlark.StringDict{
		"git": Module,
	}

	thread := &starlark.Thread{Name: "test"}
	return starlark.Eval(thread, "<test>", expr, predeclared)
}

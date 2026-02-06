package metadata_test

import (
	"strings"
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/metadata"
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

func TestSquashNotesCompact(t *testing.T) {
	ctx := transform.NewContext("/tmp")
	ctx.Changes.Current = []*transform.Change{
		{Ref: "abc123", Author: "John Doe <john@example.com>", Message: "First commit"},
		{Ref: "def456", Author: "Jane Smith <jane@example.com>", Message: "Second commit with a very long description that should be truncated"},
	}

	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.squash_notes(prefix = "Changes:\n", compact = True, max = 10)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	squash := val.(*metadata.SquashNotes)
	if err := squash.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.HasPrefix(ctx.Message, "Changes:\n") {
		t.Errorf("message should start with prefix")
	}

	if !strings.Contains(ctx.Message, "abc123") {
		t.Errorf("message should contain first commit ref")
	}
}

func TestSquashNotesNonCompact(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.squash_notes(prefix = "Changes:\n", compact = False, show_ref = True, show_author = True, show_description = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	squash := val.(*metadata.SquashNotes)

	ctx := transform.NewContext("/tmp")
	ctx.Changes.Current = []*transform.Change{
		{Ref: "abc123", Author: "John Doe <john@example.com>", Message: "First commit\n\nExtended description"},
	}

	if err := squash.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.Contains(ctx.Message, "--\n") {
		t.Errorf("non-compact format should contain '--' separator")
	}

	if !strings.Contains(ctx.Message, "Extended description") {
		t.Errorf("non-compact format should include full message")
	}
}

func TestSquashNotesOldestFirst(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.squash_notes(prefix = "Changes:\n", oldest_first = True, compact = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	squash := val.(*metadata.SquashNotes)

	ctx := transform.NewContext("/tmp")
	ctx.Changes.Current = []*transform.Change{
		{Ref: "newer", Author: "A <a@a.com>", Message: "Newer"},
		{Ref: "older", Author: "B <b@b.com>", Message: "Older"},
	}

	if err := squash.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// With oldest_first, "older" should appear before "newer"
	olderIdx := strings.Index(ctx.Message, "older")
	newerIdx := strings.Index(ctx.Message, "newer")

	if olderIdx >= newerIdx {
		t.Errorf("oldest_first should put older commits first: %s", ctx.Message)
	}
}

func TestSquashNotesUseMerge(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	// Test with use_merge = False
	val, err := starlark.Eval(thread, "test.sky", `metadata.squash_notes(prefix = "Changes:\n", use_merge = False, compact = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	squash := val.(*metadata.SquashNotes)

	ctx := transform.NewContext("/tmp")
	ctx.Changes.Current = []*transform.Change{
		{Ref: "normal", Author: "A <a@a.com>", Message: "Normal commit", IsMerge: false},
		{Ref: "merge", Author: "B <b@b.com>", Message: "Merge commit", IsMerge: true},
	}

	if err := squash.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if strings.Contains(ctx.Message, "merge") {
		t.Errorf("use_merge=False should exclude merge commits: %s", ctx.Message)
	}

	if !strings.Contains(ctx.Message, "normal") {
		t.Errorf("should include normal commits: %s", ctx.Message)
	}
}

func TestSquashNotesMaxLimit(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.squash_notes(prefix = "Changes:\n", max = 2, compact = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	squash := val.(*metadata.SquashNotes)

	ctx := transform.NewContext("/tmp")
	ctx.Changes.Current = []*transform.Change{
		{Ref: "c1", Author: "A <a@a.com>", Message: "Commit 1"},
		{Ref: "c2", Author: "A <a@a.com>", Message: "Commit 2"},
		{Ref: "c3", Author: "A <a@a.com>", Message: "Commit 3"},
		{Ref: "c4", Author: "A <a@a.com>", Message: "Commit 4"},
	}

	if err := squash.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.Contains(ctx.Message, "(And 2 more changes)") {
		t.Errorf("should indicate remaining changes: %s", ctx.Message)
	}
}

func TestScrubberWithReplacement(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	// Test extracting content from tags
	val, err := starlark.Eval(thread, "test.sky", `metadata.scrubber(regex = ".*<public>(.*)</public>.*", replacement = "$1")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scrubber := val.(*metadata.Scrubber)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "secret<public>public content</public>more secret"

	if err := scrubber.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Message != "public content" {
		t.Errorf("expected 'public content', got %q", ctx.Message)
	}
}

func TestScrubberMsgIfNoMatch(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.scrubber(regex = "PATTERN_NOT_FOUND", msg_if_no_match = "Default message")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scrubber := val.(*metadata.Scrubber)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original message without pattern"

	if err := scrubber.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Message != "Default message" {
		t.Errorf("expected 'Default message', got %q", ctx.Message)
	}
}

func TestScrubberFailIfNoMatch(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.scrubber(regex = "PATTERN_NOT_FOUND", fail_if_no_match = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scrubber := val.(*metadata.Scrubber)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original message without pattern"

	err = scrubber.Apply(ctx)
	if err == nil {
		t.Fatal("expected error when pattern doesn't match with fail_if_no_match=True")
	}
}

func TestMapAuthorByName(t *testing.T) {
	ma, err := metadata.NewMapAuthor(
		map[string]string{
			"John": "John Doe <john@new.com>",
		},
		false, false, false, false, false,
	)
	if err != nil {
		t.Fatalf("NewMapAuthor failed: %v", err)
	}

	ctx := transform.NewContext("/tmp")
	ctx.Author = "John <john@old.com>"

	if err := ma.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Author != "John Doe <john@new.com>" {
		t.Errorf("expected mapped author, got %q", ctx.Author)
	}
}

func TestMapAuthorByEmail(t *testing.T) {
	ma, err := metadata.NewMapAuthor(
		map[string]string{
			"john@old.com": "John Doe <john@new.com>",
		},
		false, false, false, false, false,
	)
	if err != nil {
		t.Fatalf("NewMapAuthor failed: %v", err)
	}

	ctx := transform.NewContext("/tmp")
	ctx.Author = "Anyone <john@old.com>"

	if err := ma.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Author != "John Doe <john@new.com>" {
		t.Errorf("expected mapped author, got %q", ctx.Author)
	}
}

func TestMapAuthorByFullAuthor(t *testing.T) {
	ma, err := metadata.NewMapAuthor(
		map[string]string{
			"John <john@old.com>": "John Doe <john@new.com>",
		},
		false, false, false, false, false,
	)
	if err != nil {
		t.Fatalf("NewMapAuthor failed: %v", err)
	}

	ctx := transform.NewContext("/tmp")
	ctx.Author = "John <john@old.com>"

	if err := ma.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Author != "John Doe <john@new.com>" {
		t.Errorf("expected mapped author, got %q", ctx.Author)
	}
}

func TestMapAuthorMapAll(t *testing.T) {
	ma, err := metadata.NewMapAuthor(
		map[string]string{
			"john@old.com": "John Doe <john@new.com>",
		},
		false, false, false, false, true, // mapAll = true
	)
	if err != nil {
		t.Fatalf("NewMapAuthor failed: %v", err)
	}

	ctx := transform.NewContext("/tmp")
	ctx.Author = "Main <main@example.com>"
	ctx.Changes.Current = []*transform.Change{
		{Author: "Someone <john@old.com>", Message: "Commit"},
	}

	if err := ma.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Changes.Current[0].MappedAuthor != "John Doe <john@new.com>" {
		t.Errorf("expected change author to be mapped, got %q", ctx.Changes.Current[0].MappedAuthor)
	}
}

func TestMapAuthorReverse(t *testing.T) {
	ma, err := metadata.NewMapAuthor(
		map[string]string{
			"John <john@old.com>": "John Doe <john@new.com>",
		},
		true, false, false, false, false, // reversible = true
	)
	if err != nil {
		t.Fatalf("NewMapAuthor failed: %v", err)
	}

	reverse := ma.Reverse()
	reverseMA, ok := reverse.(*metadata.MapAuthor)
	if !ok {
		t.Fatalf("expected *MapAuthor, got %T", reverse)
	}

	ctx := transform.NewContext("/tmp")
	ctx.Author = "John Doe <john@new.com>"

	if err := reverseMA.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Author != "John <john@old.com>" {
		t.Errorf("expected reverse mapped author, got %q", ctx.Author)
	}
}

func TestExposeLabelAll(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.expose_label(name = "LABEL", all = True, separator = "=")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expose := val.(*metadata.ExposeLabel)

	ctx := transform.NewContext("/tmp")
	ctx.Labels["LABEL"] = []string{"value1", "value2"}
	ctx.Message = "Original"

	if err := expose.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should have both values in message
	if !strings.Contains(ctx.Message, "LABEL=value1") || !strings.Contains(ctx.Message, "LABEL=value2") {
		t.Errorf("expected both label values in message: %s", ctx.Message)
	}
}

func TestExposeLabelAllWithConcat(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.expose_label(name = "LABEL", all = True, concat_separator = ",", separator = "=")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expose := val.(*metadata.ExposeLabel)

	ctx := transform.NewContext("/tmp")
	ctx.Labels["LABEL"] = []string{"value1", "value2"}
	ctx.Message = "Original"

	if err := expose.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should have concatenated values
	if !strings.Contains(ctx.Message, "LABEL=value1,value2") {
		t.Errorf("expected concatenated label values in message: %s", ctx.Message)
	}
}

func TestAddHeaderNoNewLine(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.add_header(text = "PREFIX: ", new_line = False)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := val.(*metadata.AddHeader)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original message"

	if err := header.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	expected := "PREFIX: Original message"
	if ctx.Message != expected {
		t.Errorf("expected %q, got %q", expected, ctx.Message)
	}
}

func TestAddHeaderWithTemplate(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.add_header(text = "Issue: ${ISSUE}", new_line = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	header := val.(*metadata.AddHeader)

	ctx := transform.NewContext("/tmp")
	ctx.Labels["ISSUE"] = []string{"12345"}
	ctx.Message = "Original message"

	if err := header.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !strings.HasPrefix(ctx.Message, "Issue: 12345\n") {
		t.Errorf("expected header with resolved template: %q", ctx.Message)
	}
}

func TestNoopTransformation(t *testing.T) {
	noop := transform.NewNoopTransformation(nil)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original"

	if err := noop.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Message != "Original" {
		t.Errorf("noop should not change message")
	}

	if noop.Describe() != "noop" {
		t.Errorf("unexpected description: %s", noop.Describe())
	}
}

func TestTransformationDescribe(t *testing.T) {
	tests := []struct {
		name         string
		transform    transform.Transformation
		wantDescribe string
	}{
		{
			name:         "SquashNotes",
			transform:    &metadata.SquashNotes{},
			wantDescribe: "squash_notes",
		},
		{
			name:         "RestoreAuthor",
			transform:    &metadata.RestoreAuthor{},
			wantDescribe: "Restoring original author",
		},
		{
			name:         "ReplaceMessage",
			transform:    &metadata.ReplaceMessage{},
			wantDescribe: "Replacing commit message",
		},
		{
			name:         "AddHeader",
			transform:    &metadata.AddHeader{},
			wantDescribe: "Adding header to commit message",
		},
		{
			name:         "Scrubber",
			transform:    &metadata.Scrubber{},
			wantDescribe: "Description scrubber",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.transform.Describe(); got != tt.wantDescribe {
				t.Errorf("Describe() = %q, want %q", got, tt.wantDescribe)
			}
		})
	}
}

package metadata_test

import (
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/metadata"
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

func TestModule(t *testing.T) {
	if metadata.Module == nil {
		t.Fatal("expected non-nil module")
	}

	if metadata.Module.Name != "metadata" {
		t.Errorf("expected module name 'metadata', got %q", metadata.Module.Name)
	}
}

func TestSquashNotes(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "default parameters",
			code:    `metadata.squash_notes()`,
			wantErr: false,
		},
		{
			name:    "with prefix",
			code:    `metadata.squash_notes(prefix = "Changes:\n")`,
			wantErr: false,
		},
		{
			name:    "with all parameters",
			code:    `metadata.squash_notes(prefix = "Import:\n", max = 50, compact = False, show_ref = True, show_author = False, show_description = True, oldest_first = True, use_merge = False)`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			sn, ok := val.(*metadata.SquashNotes)
			if !ok {
				t.Fatalf("expected *SquashNotes, got %T", val)
			}

			if sn.String() != "metadata.squash_notes()" {
				t.Errorf("unexpected string: %s", sn.String())
			}
		})
	}
}

func TestSquashNotesApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.squash_notes(prefix = "Changes:\n\n", compact = True, show_ref = True, show_author = True, show_description = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sn := val.(*metadata.SquashNotes)

	ctx := transform.NewContext("/tmp")
	ctx.Changes.Current = []*transform.Change{
		{Ref: "abc123", Author: "John Doe <john@example.com>", Message: "First commit\n\nDetails here"},
		{Ref: "def456", Author: "Jane Smith <jane@example.com>", Message: "Second commit"},
	}

	if err := sn.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	expected := "Changes:\n\n  - abc123 First commit by John Doe <john@example.com>\n  - def456 Second commit by Jane Smith <jane@example.com>\n"
	if ctx.Message != expected {
		t.Errorf("unexpected message:\ngot:  %q\nwant: %q", ctx.Message, expected)
	}
}

func TestSaveAuthor(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "default label",
			code:    `metadata.save_author()`,
			wantErr: false,
		},
		{
			name:    "custom label",
			code:    `metadata.save_author(label = "AUTHOR")`,
			wantErr: false,
		},
		{
			name:    "custom separator",
			code:    `metadata.save_author(separator = ": ")`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			_, ok := val.(*metadata.SaveAuthor)
			if !ok {
				t.Fatalf("expected *SaveAuthor, got %T", val)
			}
		})
	}
}

func TestSaveAuthorApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.save_author(label = "AUTHOR", separator = "=")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sa := val.(*metadata.SaveAuthor)

	ctx := transform.NewContext("/tmp")
	ctx.Author = "John Doe <john@example.com>"
	ctx.Message = "Initial commit"

	if err := sa.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.GetLabel("AUTHOR") != "John Doe <john@example.com>" {
		t.Errorf("expected author label to be set, got %q", ctx.GetLabel("AUTHOR"))
	}
}

func TestRestoreAuthor(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "default parameters",
			code:    `metadata.restore_author()`,
			wantErr: false,
		},
		{
			name:    "custom label",
			code:    `metadata.restore_author(label = "AUTHOR")`,
			wantErr: false,
		},
		{
			name:    "search all changes",
			code:    `metadata.restore_author(search_all_changes = True)`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			_, ok := val.(*metadata.RestoreAuthor)
			if !ok {
				t.Fatalf("expected *RestoreAuthor, got %T", val)
			}
		})
	}
}

func TestRestoreAuthorApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.restore_author(label = "ORIGINAL_AUTHOR")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ra := val.(*metadata.RestoreAuthor)

	ctx := transform.NewContext("/tmp")
	ctx.Author = "Default Author <default@example.com>"
	ctx.Message = "Commit message\nORIGINAL_AUTHOR=John Doe <john@example.com>\n"

	if err := ra.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Author != "John Doe <john@example.com>" {
		t.Errorf("expected author to be restored, got %q", ctx.Author)
	}
}

func TestReplaceMessage(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "simple text",
			code:    `metadata.replace_message(text = "New message")`,
			wantErr: false,
		},
		{
			name:    "with template",
			code:    `metadata.replace_message(text = "Import of ${LABEL}")`,
			wantErr: false,
		},
		{
			name:    "ignore label not found",
			code:    `metadata.replace_message(text = "Import of ${LABEL}", ignore_label_not_found = True)`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if err != nil {
				return
			}

			_, ok := val.(*metadata.ReplaceMessage)
			if !ok {
				t.Fatalf("expected *ReplaceMessage, got %T", val)
			}
		})
	}
}

func TestReplaceMessageApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.replace_message(text = "Import: ${ISSUE}")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rm := val.(*metadata.ReplaceMessage)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original message\nISSUE=12345"

	if err := rm.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Message != "Import: 12345" {
		t.Errorf("unexpected message: %q", ctx.Message)
	}
}

func TestExposeLabel(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "simple label",
			code:    `metadata.expose_label(name = "REVIEW_URL")`,
			wantErr: false,
		},
		{
			name:    "with new name",
			code:    `metadata.expose_label(name = "REVIEW_URL", new_name = "URL")`,
			wantErr: false,
		},
		{
			name:    "custom separator",
			code:    `metadata.expose_label(name = "REVIEW_URL", separator = ": ")`,
			wantErr: false,
		},
		{
			name:    "expose all",
			code:    `metadata.expose_label(name = "REVIEW_URL", all = True)`,
			wantErr: false,
		},
		{
			name:    "expose all with concat",
			code:    `metadata.expose_label(name = "REVIEW_URL", all = True, concat_separator = ",")`,
			wantErr: false,
		},
		{
			name:    "invalid label name",
			code:    `metadata.expose_label(name = "123invalid")`,
			wantErr: true,
		},
		{
			name:    "concat without all",
			code:    `metadata.expose_label(name = "LABEL", concat_separator = ",")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestExposeLabelApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.expose_label(name = "HIDDEN_LABEL", new_name = "EXPOSED_LABEL", separator = "=")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	el := val.(*metadata.ExposeLabel)

	ctx := transform.NewContext("/tmp")
	ctx.Labels["HIDDEN_LABEL"] = []string{"secret_value"}
	ctx.Message = "Original message"

	if err := el.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.GetLabel("EXPOSED_LABEL") != "secret_value" {
		t.Errorf("expected exposed label, got %q", ctx.GetLabel("EXPOSED_LABEL"))
	}
}

func TestAddHeader(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "simple header",
			code:    `metadata.add_header(text = "COPYBARA CHANGE")`,
			wantErr: false,
		},
		{
			name:    "with template",
			code:    `metadata.add_header(text = "Import of ${LABEL}")`,
			wantErr: false,
		},
		{
			name:    "no new line",
			code:    `metadata.add_header(text = "PREFIX: ", new_line = False)`,
			wantErr: false,
		},
		{
			name:    "ignore label not found",
			code:    `metadata.add_header(text = "${LABEL}", ignore_label_not_found = True)`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAddHeaderApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.add_header(text = "COPYBARA CHANGE", new_line = True)`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ah := val.(*metadata.AddHeader)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original message"

	if err := ah.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	expected := "COPYBARA CHANGE\nOriginal message"
	if ctx.Message != expected {
		t.Errorf("unexpected message:\ngot:  %q\nwant: %q", ctx.Message, expected)
	}
}

func TestScrubber(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "single regex",
			code:    `metadata.scrubber(regex = "CONFIDENTIAL:.*")`,
			wantErr: false,
		},
		{
			name:    "multiple regexes",
			code:    `metadata.scrubber(regexes = ["SECRET:.*", "PRIVATE:.*"])`,
			wantErr: false,
		},
		{
			name:    "with replacement",
			code:    `metadata.scrubber(regex = "<public>(.*)</public>", replacement = "$1")`,
			wantErr: false,
		},
		{
			name:    "with default message",
			code:    `metadata.scrubber(regex = "pattern", msg_if_no_match = "Default message")`,
			wantErr: false,
		},
		{
			name:    "fail if no match",
			code:    `metadata.scrubber(regex = "pattern", fail_if_no_match = True)`,
			wantErr: false,
		},
		{
			name:    "no regex provided",
			code:    `metadata.scrubber()`,
			wantErr: true,
		},
		{
			name:    "fail and default message conflict",
			code:    `metadata.scrubber(regex = "pattern", fail_if_no_match = True, msg_if_no_match = "Default")`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestScrubberApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.scrubber(regex = "CONFIDENTIAL:.*", replacement = "")`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scrubber := val.(*metadata.Scrubber)

	ctx := transform.NewContext("/tmp")
	ctx.Message = "Public info\nCONFIDENTIAL: secret stuff\nMore public"

	if err := scrubber.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	expected := "Public info\n\nMore public"
	if ctx.Message != expected {
		t.Errorf("unexpected message:\ngot:  %q\nwant: %q", ctx.Message, expected)
	}
}

func TestMapAuthor(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name:    "simple mapping",
			code:    `metadata.map_author(authors = {"john@old.com": "John Doe <john@new.com>"})`,
			wantErr: false,
		},
		{
			name:    "reversible mapping",
			code:    `metadata.map_author(authors = {"John <john@old.com>": "John Doe <john@new.com>"}, reversible = True)`,
			wantErr: false,
		},
		{
			name:    "noop reverse",
			code:    `metadata.map_author(authors = {"john": "John Doe <john@new.com>"}, noop_reverse = True)`,
			wantErr: false,
		},
		{
			name:    "fail if not found",
			code:    `metadata.map_author(authors = {"john": "John Doe <john@new.com>"}, fail_if_not_found = True)`,
			wantErr: false,
		},
		{
			name:    "map all changes",
			code:    `metadata.map_author(authors = {"john": "John Doe <john@new.com>"}, map_all_changes = True)`,
			wantErr: false,
		},
		{
			name:    "reverse_fail without reversible",
			code:    `metadata.map_author(authors = {}, reverse_fail_if_not_found = True)`,
			wantErr: true,
		},
		{
			name:    "reverse_fail with noop_reverse",
			code:    `metadata.map_author(authors = {}, reversible = True, noop_reverse = True, reverse_fail_if_not_found = True)`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := starlark.Eval(thread, "test.sky", tt.code, predeclared)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMapAuthorApply(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `metadata.map_author(authors = {"john@old.com": "John Doe <john@new.com>"})`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ma := val.(*metadata.MapAuthor)

	ctx := transform.NewContext("/tmp")
	ctx.Author = "John <john@old.com>"

	if err := ma.Apply(ctx); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if ctx.Author != "John Doe <john@new.com>" {
		t.Errorf("unexpected author: %q", ctx.Author)
	}
}

func TestTransformationReverse(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"metadata": metadata.Module,
	}

	// Test SaveAuthor.Reverse() returns RestoreAuthor
	saveVal, _ := starlark.Eval(thread, "test.sky", `metadata.save_author(label = "AUTHOR")`, predeclared)
	save := saveVal.(*metadata.SaveAuthor)
	reverse := save.Reverse()
	if _, ok := reverse.(*metadata.RestoreAuthor); !ok {
		t.Errorf("SaveAuthor.Reverse() should return RestoreAuthor, got %T", reverse)
	}

	// Test RestoreAuthor.Reverse() returns SaveAuthor
	restoreVal, _ := starlark.Eval(thread, "test.sky", `metadata.restore_author(label = "AUTHOR")`, predeclared)
	restore := restoreVal.(*metadata.RestoreAuthor)
	reverse = restore.Reverse()
	if _, ok := reverse.(*metadata.SaveAuthor); !ok {
		t.Errorf("RestoreAuthor.Reverse() should return SaveAuthor, got %T", reverse)
	}

	// Test SquashNotes.Reverse() returns NoopTransformation
	squashVal, _ := starlark.Eval(thread, "test.sky", `metadata.squash_notes()`, predeclared)
	squash := squashVal.(*metadata.SquashNotes)
	reverse = squash.Reverse()
	if reverse.Describe() != "noop" {
		t.Errorf("SquashNotes.Reverse().Describe() should be 'noop', got %q", reverse.Describe())
	}
}

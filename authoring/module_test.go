package authoring_test

import (
	"testing"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/authoring"
)

func TestModule(t *testing.T) {
	if authoring.Module == nil {
		t.Fatal("expected non-nil module")
	}

	if authoring.Module.Name != "authoring" {
		t.Errorf("expected module name 'authoring', got %q", authoring.Module.Name)
	}

	// Check all expected members exist
	members := []string{"pass_thru", "overwrite", "allowed", "new_author"}
	for _, name := range members {
		if _, ok := authoring.Module.Members[name]; !ok {
			t.Errorf("expected member %q not found", name)
		}
	}
}

func TestPassThru(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`authoring.pass_thru(default = "Foo Bar <foo@bar.com>")`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pt, ok := val.(*authoring.PassThru)
	if !ok {
		t.Fatalf("expected *PassThru, got %T", val)
	}

	if pt.Type() != "authoring.pass_thru" {
		t.Errorf("Type() = %q, want %q", pt.Type(), "authoring.pass_thru")
	}

	if pt.Mode() != authoring.ModePassThru {
		t.Errorf("Mode() = %v, want ModePassThru", pt.Mode())
	}

	if pt.DefaultAuthor().Email() != "foo@bar.com" {
		t.Errorf("DefaultAuthor().Email() = %q, want %q", pt.DefaultAuthor().Email(), "foo@bar.com")
	}
}

func TestPassThru_MissingDefault(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	_, err := starlark.Eval(thread, "test.sky",
		`authoring.pass_thru()`,
		predeclared,
	)
	if err == nil {
		t.Fatal("expected error for missing default parameter")
	}
}

func TestPassThru_InvalidAuthor(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	_, err := starlark.Eval(thread, "test.sky",
		`authoring.pass_thru(default = "invalid")`,
		predeclared,
	)
	if err == nil {
		t.Fatal("expected error for invalid author format")
	}
}

func TestPassThru_ResolveAuthor(t *testing.T) {
	defaultAuthor := authoring.NewAuthor("Default", "default@example.com")
	originalAuthor := authoring.NewAuthor("Original", "original@example.com")

	pt := &authoring.PassThru{}
	// Use reflection or create via Starlark to set defaultAuthor
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`authoring.pass_thru(default = "Default <default@example.com>")`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pt = val.(*authoring.PassThru)

	// Test with original author
	resolved := pt.ResolveAuthor(originalAuthor)
	if resolved.Email() != "original@example.com" {
		t.Errorf("ResolveAuthor(original) = %s, want original author", resolved)
	}

	// Test with nil author
	resolved = pt.ResolveAuthor(nil)
	if resolved.Email() != defaultAuthor.Email() {
		t.Errorf("ResolveAuthor(nil) = %s, want default author", resolved)
	}
}

func TestOverwrite(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`authoring.overwrite(default = "Foo Bar <noreply@foobar.com>")`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ow, ok := val.(*authoring.Overwrite)
	if !ok {
		t.Fatalf("expected *Overwrite, got %T", val)
	}

	if ow.Type() != "authoring.overwrite" {
		t.Errorf("Type() = %q, want %q", ow.Type(), "authoring.overwrite")
	}

	if ow.Mode() != authoring.ModeOverwrite {
		t.Errorf("Mode() = %v, want ModeOverwrite", ow.Mode())
	}

	if ow.DefaultAuthor().Email() != "noreply@foobar.com" {
		t.Errorf("DefaultAuthor().Email() = %q, want %q", ow.DefaultAuthor().Email(), "noreply@foobar.com")
	}
}

func TestOverwrite_ResolveAuthor(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`authoring.overwrite(default = "Bot <bot@example.com>")`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ow := val.(*authoring.Overwrite)
	originalAuthor := authoring.NewAuthor("Original", "original@example.com")

	// Should always return the configured author
	resolved := ow.ResolveAuthor(originalAuthor)
	if resolved.Email() != "bot@example.com" {
		t.Errorf("ResolveAuthor(original) = %s, want bot author", resolved)
	}

	resolved = ow.ResolveAuthor(nil)
	if resolved.Email() != "bot@example.com" {
		t.Errorf("ResolveAuthor(nil) = %s, want bot author", resolved)
	}
}

func TestAllowed(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `
authoring.allowed(
    default = "Default <default@example.com>",
    allowlist = [
        "user1@example.com",
        "user2@example.com",
    ],
)`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, ok := val.(*authoring.Allowed)
	if !ok {
		t.Fatalf("expected *Allowed, got %T", val)
	}

	if allowed.Type() != "authoring.allowed" {
		t.Errorf("Type() = %q, want %q", allowed.Type(), "authoring.allowed")
	}

	if allowed.Mode() != authoring.ModeAllowed {
		t.Errorf("Mode() = %v, want ModeAllowed", allowed.Mode())
	}

	if allowed.DefaultAuthor().Email() != "default@example.com" {
		t.Errorf("DefaultAuthor().Email() = %q, want %q", allowed.DefaultAuthor().Email(), "default@example.com")
	}

	allowlist := allowed.Allowlist()
	if len(allowlist) != 2 {
		t.Errorf("Allowlist() len = %d, want 2", len(allowlist))
	}
}

func TestAllowed_EmptyAllowlist(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	_, err := starlark.Eval(thread, "test.sky",
		`authoring.allowed(default = "Default <default@example.com>", allowlist = [])`,
		predeclared,
	)
	if err == nil {
		t.Fatal("expected error for empty allowlist")
	}
}

func TestAllowed_DuplicateEntry(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	_, err := starlark.Eval(thread, "test.sky", `
authoring.allowed(
    default = "Default <default@example.com>",
    allowlist = ["user@example.com", "user@example.com"],
)`,
		predeclared,
	)
	if err == nil {
		t.Fatal("expected error for duplicate allowlist entry")
	}
}

func TestAllowed_ResolveAuthor(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `
authoring.allowed(
    default = "Default <default@example.com>",
    allowlist = [
        "allowed@example.com",
        "User Two <user2@example.com>",
    ],
)`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed := val.(*authoring.Allowed)

	tests := []struct {
		name       string
		author     *authoring.Author
		wantEmail  string
		wantReason string
	}{
		{
			name:       "allowed by email",
			author:     authoring.NewAuthor("Any Name", "allowed@example.com"),
			wantEmail:  "allowed@example.com",
			wantReason: "should pass through allowed email",
		},
		{
			name:       "allowed by full string",
			author:     authoring.NewAuthor("User Two", "user2@example.com"),
			wantEmail:  "user2@example.com",
			wantReason: "should pass through allowed full string",
		},
		{
			name:       "not allowed",
			author:     authoring.NewAuthor("Random", "random@example.com"),
			wantEmail:  "default@example.com",
			wantReason: "should use default for non-allowed",
		},
		{
			name:       "nil author",
			author:     nil,
			wantEmail:  "default@example.com",
			wantReason: "should use default for nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := allowed.ResolveAuthor(tt.author)
			if resolved.Email() != tt.wantEmail {
				t.Errorf("ResolveAuthor() = %s, %s", resolved, tt.wantReason)
			}
		})
	}
}

func TestAllowed_RegexPatterns(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky", `
authoring.allowed(
    default = "Default <default@example.com>",
    allowlist = [
        "/.*@myorg\\.com$/",
        "/^team-.*@/",
    ],
)`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed := val.(*authoring.Allowed)

	tests := []struct {
		name      string
		author    *authoring.Author
		wantEmail string
	}{
		{
			name:      "matches org pattern",
			author:    authoring.NewAuthor("Employee", "employee@myorg.com"),
			wantEmail: "employee@myorg.com",
		},
		{
			name:      "matches team pattern",
			author:    authoring.NewAuthor("Team Lead", "team-backend@example.com"),
			wantEmail: "team-backend@example.com",
		},
		{
			name:      "no match",
			author:    authoring.NewAuthor("External", "external@other.com"),
			wantEmail: "default@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := allowed.ResolveAuthor(tt.author)
			if resolved.Email() != tt.wantEmail {
				t.Errorf("ResolveAuthor(%s) = %s, want email %s", tt.author, resolved, tt.wantEmail)
			}
		})
	}
}

func TestAllowed_InvalidRegex(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	_, err := starlark.Eval(thread, "test.sky", `
authoring.allowed(
    default = "Default <default@example.com>",
    allowlist = ["/[invalid/"],
)`,
		predeclared,
	)
	if err == nil {
		t.Fatal("expected error for invalid regex pattern")
	}
}

func TestNewAuthorFn(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	val, err := starlark.Eval(thread, "test.sky",
		`authoring.new_author(name = "John Doe", email = "john@example.com")`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	author, ok := val.(*authoring.Author)
	if !ok {
		t.Fatalf("expected *Author, got %T", val)
	}

	if author.Name() != "John Doe" {
		t.Errorf("Name() = %q, want %q", author.Name(), "John Doe")
	}

	if author.Email() != "john@example.com" {
		t.Errorf("Email() = %q, want %q", author.Email(), "john@example.com")
	}
}

func TestNewAuthorFn_EmptyName(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	_, err := starlark.Eval(thread, "test.sky",
		`authoring.new_author(name = "", email = "john@example.com")`,
		predeclared,
	)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNewAuthorFn_EmptyEmail(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	// Empty email should be allowed
	val, err := starlark.Eval(thread, "test.sky",
		`authoring.new_author(name = "Bot", email = "")`,
		predeclared,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	author := val.(*authoring.Author)
	if author.Email() != "" {
		t.Errorf("Email() = %q, want empty", author.Email())
	}
}

func TestAuthoringMode_String(t *testing.T) {
	tests := []struct {
		mode authoring.AuthoringMode
		want string
	}{
		{authoring.ModePassThru, "PASS_THRU"},
		{authoring.ModeOverwrite, "OVERWRITE"},
		{authoring.ModeAllowed, "ALLOWED"},
		{authoring.AuthoringMode(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStarlarkIntegration(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	// Test accessing author attributes in Starlark
	// Use ExecFile instead of Eval for multi-statement code
	globals, err := starlark.ExecFile(thread, "test.sky", `
author = authoring.new_author(name = "John Doe", email = "john@example.com")
result = author.name + " - " + author.email
`, predeclared)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := globals["result"]
	if result.String() != `"John Doe - john@example.com"` {
		t.Errorf("got %s, want %q", result, "John Doe - john@example.com")
	}
}

func TestValueFreeze(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	// Create values
	ptVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.pass_thru(default = "A <a@b.com>")`, predeclared)
	owVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.overwrite(default = "A <a@b.com>")`, predeclared)
	alVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.allowed(default = "A <a@b.com>", allowlist = ["x@y.com"])`, predeclared)
	auVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.new_author(name = "A", email = "a@b.com")`, predeclared)

	// Freeze should not panic
	ptVal.Freeze()
	owVal.Freeze()
	alVal.Freeze()
	auVal.Freeze()
}

func TestValueHash(t *testing.T) {
	thread := &starlark.Thread{Name: "test"}
	predeclared := starlark.StringDict{
		"authoring": authoring.Module,
	}

	// These should return errors (unhashable)
	ptVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.pass_thru(default = "A <a@b.com>")`, predeclared)
	owVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.overwrite(default = "A <a@b.com>")`, predeclared)
	alVal, _ := starlark.Eval(thread, "test.sky",
		`authoring.allowed(default = "A <a@b.com>", allowlist = ["x@y.com"])`, predeclared)

	_, err := ptVal.(*authoring.PassThru).Hash()
	if err == nil {
		t.Error("PassThru.Hash() expected error")
	}

	_, err = owVal.(*authoring.Overwrite).Hash()
	if err == nil {
		t.Error("Overwrite.Hash() expected error")
	}

	_, err = alVal.(*authoring.Allowed).Hash()
	if err == nil {
		t.Error("Allowed.Hash() expected error")
	}
}

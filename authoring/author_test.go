package authoring_test

import (
	"errors"
	"testing"

	"github.com/albertocavalcante/starlark-go-copybara/authoring"
)

func TestParseAuthor(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantEmail string
		wantErr   bool
	}{
		{
			name:      "standard format",
			input:     "John Doe <john@example.com>",
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name:      "with extra spaces",
			input:     "  John Doe  <  john@example.com  >",
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name:      "double quoted",
			input:     `"John Doe <john@example.com>"`,
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name:      "single quoted",
			input:     `'John Doe <john@example.com>'`,
			wantName:  "John Doe",
			wantEmail: "john@example.com",
		},
		{
			name:      "empty email",
			input:     "John Doe <>",
			wantName:  "John Doe",
			wantEmail: "",
		},
		{
			name:      "name with special characters",
			input:     "John O'Brien <john@example.com>",
			wantName:  "John O'Brien",
			wantEmail: "john@example.com",
		},
		{
			name:      "name with numbers",
			input:     "Bot 123 <bot@example.com>",
			wantName:  "Bot 123",
			wantEmail: "bot@example.com",
		},
		{
			name:    "missing angle brackets",
			input:   "John Doe john@example.com",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only email",
			input:   "<john@example.com>",
			wantErr: true,
		},
		{
			name:    "missing closing bracket",
			input:   "John Doe <john@example.com",
			wantErr: true,
		},
		{
			name:    "empty name with spaces",
			input:   "   <john@example.com>",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			author, err := authoring.ParseAuthor(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseAuthor(%q) expected error, got nil", tt.input)
				}
				if !errors.Is(err, authoring.ErrInvalidAuthor) {
					t.Errorf("ParseAuthor(%q) error = %v, want ErrInvalidAuthor", tt.input, err)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseAuthor(%q) unexpected error: %v", tt.input, err)
				return
			}

			if author.Name() != tt.wantName {
				t.Errorf("ParseAuthor(%q).Name() = %q, want %q", tt.input, author.Name(), tt.wantName)
			}

			if author.Email() != tt.wantEmail {
				t.Errorf("ParseAuthor(%q).Email() = %q, want %q", tt.input, author.Email(), tt.wantEmail)
			}
		})
	}
}

func TestNewAuthor(t *testing.T) {
	author := authoring.NewAuthor("John Doe", "john@example.com")

	if author.Name() != "John Doe" {
		t.Errorf("Name() = %q, want %q", author.Name(), "John Doe")
	}

	if author.Email() != "john@example.com" {
		t.Errorf("Email() = %q, want %q", author.Email(), "john@example.com")
	}
}

func TestAuthorString(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{
			name:  "John Doe",
			email: "john@example.com",
			want:  "John Doe <john@example.com>",
		},
		{
			name:  "Bot",
			email: "",
			want:  "Bot <>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			author := authoring.NewAuthor(tt.name, tt.email)
			if got := author.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAuthorType(t *testing.T) {
	author := authoring.NewAuthor("John", "john@example.com")
	if got := author.Type(); got != "author" {
		t.Errorf("Type() = %q, want %q", got, "author")
	}
}

func TestAuthorTruth(t *testing.T) {
	author := authoring.NewAuthor("John", "john@example.com")
	if !author.Truth() {
		t.Error("Truth() = false, want true")
	}
}

func TestAuthorHash(t *testing.T) {
	author1 := authoring.NewAuthor("John", "john@example.com")
	author2 := authoring.NewAuthor("Jane", "john@example.com") // Same email
	author3 := authoring.NewAuthor("John", "jane@example.com") // Different email

	hash1, err1 := author1.Hash()
	hash2, err2 := author2.Hash()
	hash3, err3 := author3.Hash()

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("Hash() returned unexpected error")
	}

	// Same email should have same hash
	if hash1 != hash2 {
		t.Errorf("Authors with same email should have same hash: %d != %d", hash1, hash2)
	}

	// Different email should (likely) have different hash
	if hash1 == hash3 {
		t.Errorf("Authors with different emails should have different hash: %d == %d", hash1, hash3)
	}
}

func TestAuthorHashEmptyEmail(t *testing.T) {
	author1 := authoring.NewAuthor("John", "")
	author2 := authoring.NewAuthor("John", "")
	author3 := authoring.NewAuthor("Jane", "")

	hash1, _ := author1.Hash()
	hash2, _ := author2.Hash()
	hash3, _ := author3.Hash()

	// Same name (with empty email) should have same hash
	if hash1 != hash2 {
		t.Errorf("Authors with same name should have same hash: %d != %d", hash1, hash2)
	}

	// Different name should have different hash
	if hash1 == hash3 {
		t.Errorf("Authors with different names should have different hash: %d == %d", hash1, hash3)
	}
}

func TestAuthorEquals(t *testing.T) {
	tests := []struct {
		name   string
		a      *authoring.Author
		b      *authoring.Author
		equals bool
	}{
		{
			name:   "same email different name",
			a:      authoring.NewAuthor("John", "john@example.com"),
			b:      authoring.NewAuthor("Jane", "john@example.com"),
			equals: true,
		},
		{
			name:   "different email same name",
			a:      authoring.NewAuthor("John", "john@example.com"),
			b:      authoring.NewAuthor("John", "jane@example.com"),
			equals: false,
		},
		{
			name:   "same email same name",
			a:      authoring.NewAuthor("John", "john@example.com"),
			b:      authoring.NewAuthor("John", "john@example.com"),
			equals: true,
		},
		{
			name:   "empty email same name",
			a:      authoring.NewAuthor("John", ""),
			b:      authoring.NewAuthor("John", ""),
			equals: true,
		},
		{
			name:   "empty email different name",
			a:      authoring.NewAuthor("John", ""),
			b:      authoring.NewAuthor("Jane", ""),
			equals: false,
		},
		{
			name:   "one empty email",
			a:      authoring.NewAuthor("John", "john@example.com"),
			b:      authoring.NewAuthor("John", ""),
			equals: false,
		},
		{
			name:   "nil comparison",
			a:      authoring.NewAuthor("John", "john@example.com"),
			b:      nil,
			equals: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(tt.b); got != tt.equals {
				t.Errorf("Equals() = %v, want %v", got, tt.equals)
			}
		})
	}
}

func TestAuthorAttr(t *testing.T) {
	author := authoring.NewAuthor("John Doe", "john@example.com")

	name, err := author.Attr("name")
	if err != nil {
		t.Errorf("Attr(name) unexpected error: %v", err)
	}
	if name.String() != `"John Doe"` {
		t.Errorf("Attr(name) = %s, want %q", name, "John Doe")
	}

	email, err := author.Attr("email")
	if err != nil {
		t.Errorf("Attr(email) unexpected error: %v", err)
	}
	if email.String() != `"john@example.com"` {
		t.Errorf("Attr(email) = %s, want %q", email, "john@example.com")
	}

	// For unknown attributes, return nil, nil (Go starlark convention)
	val, err := author.Attr("invalid")
	if err != nil {
		t.Errorf("Attr(invalid) unexpected error: %v", err)
	}
	if val != nil {
		t.Errorf("Attr(invalid) = %v, want nil", val)
	}
}

func TestAuthorAttrNames(t *testing.T) {
	author := authoring.NewAuthor("John", "john@example.com")
	names := author.AttrNames()

	if len(names) != 2 {
		t.Errorf("AttrNames() len = %d, want 2", len(names))
	}

	// Should be sorted
	if names[0] != "email" || names[1] != "name" {
		t.Errorf("AttrNames() = %v, want [email, name]", names)
	}
}

func TestValidateAuthor(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"John Doe <john@example.com>", false},
		{"Bot <>", false},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			err := authoring.ValidateAuthor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAuthor(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"john@example.com", false},
		{"user@domain.org", false},
		{"", false}, // Empty is allowed (lenient)
		{"invalid", true},
		{"@example.com", true},
		{"john@", true},
		{"john@example", true},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := authoring.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

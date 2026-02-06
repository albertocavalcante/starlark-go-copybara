package types

import (
	"slices"
	"testing"
	"time"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/authoring"
)

// TestPathStarlarkValue tests that Path implements starlark.Value correctly.
func TestPathStarlarkValue(t *testing.T) {
	p := NewPath("/foo/bar/baz.txt")

	// Test String()
	if got := p.String(); got != "/foo/bar/baz.txt" {
		t.Errorf("String() = %q, want %q", got, "/foo/bar/baz.txt")
	}

	// Test Type()
	if got := p.Type(); got != "path" {
		t.Errorf("Type() = %q, want %q", got, "path")
	}

	// Test Truth()
	if got := p.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	// Note: NewPath("") becomes "." after filepath.Clean, which is truthy
	emptyPath := NewPath("")
	if got := emptyPath.Truth(); got != starlark.True {
		t.Errorf("empty path (becomes '.') Truth() = %v, want True", got)
	}

	// Test Hash()
	hash1, err := p.Hash()
	if err != nil {
		t.Errorf("Hash() error = %v", err)
	}
	hash2, err := NewPath("/foo/bar/baz.txt").Hash()
	if err != nil {
		t.Errorf("Hash() error = %v", err)
	}
	if hash1 != hash2 {
		t.Errorf("Hash() not equal for same paths")
	}

	// Test Freeze (should not panic)
	p.Freeze()
}

// TestPathMethods tests Path methods.
func TestPathMethods(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		testFunc func(*Path) any
		want     any
	}{
		{
			name:     "Base returns filename",
			path:     "/foo/bar/baz.txt",
			testFunc: func(p *Path) any { return p.Base() },
			want:     "baz.txt",
		},
		{
			name:     "Parent returns directory",
			path:     "/foo/bar/baz.txt",
			testFunc: func(p *Path) any { return p.Parent().String() },
			want:     "/foo/bar",
		},
		{
			name:     "Join appends segments",
			path:     "/foo",
			testFunc: func(p *Path) any { return p.Join("bar", "baz").String() },
			want:     "/foo/bar/baz",
		},
		{
			name:     "ResolveSibling replaces filename",
			path:     "/foo/bar/old.txt",
			testFunc: func(p *Path) any { return p.ResolveSibling("new.txt").String() },
			want:     "/foo/bar/new.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPath(tt.path)
			got := tt.testFunc(p)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPathStartsWith tests the StartsWith method.
func TestPathStartsWith(t *testing.T) {
	tests := []struct {
		path   string
		prefix string
		want   bool
	}{
		{"/foo/bar/baz", "/foo", true},
		{"/foo/bar/baz", "/foo/bar", true},
		{"/foo/bar/baz", "/foo/bar/baz", true},
		{"/foo/bar/baz", "/foo/ba", false}, // Not at path boundary
		{"/foo/bar", "/foo/bar/baz", false},
		{"foo/bar", "foo", true},
		{"foobar", "foo", false}, // Not at path boundary
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.prefix, func(t *testing.T) {
			p := NewPath(tt.path)
			if got := p.StartsWith(tt.prefix); got != tt.want {
				t.Errorf("StartsWith(%q) = %v, want %v", tt.prefix, got, tt.want)
			}
		})
	}
}

// TestPathRelativize tests the Relativize method.
func TestPathRelativize(t *testing.T) {
	tests := []struct {
		base  string
		other string
		want  string
	}{
		{"/foo", "/foo/bar/baz", "bar/baz"},
		{"/foo/bar", "/foo/bar/baz", "baz"},
		{"/foo", "/foo", "."},
	}

	for _, tt := range tests {
		t.Run(tt.base+"_"+tt.other, func(t *testing.T) {
			base := NewPath(tt.base)
			other := NewPath(tt.other)
			got, err := base.Relativize(other)
			if err != nil {
				t.Errorf("Relativize() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("Relativize() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// TestPathHasAttrs tests that Path implements starlark.HasAttrs.
func TestPathHasAttrs(t *testing.T) {
	p := NewPath("/foo/bar/baz.txt")

	// Test AttrNames returns sorted list
	names := p.AttrNames()
	expectedNames := []string{"join", "name", "parent", "path", "relativize", "resolve_sibling", "starts_with"}
	if !slices.Equal(names, expectedNames) {
		t.Errorf("AttrNames() = %v, want %v", names, expectedNames)
	}

	// Test Attr returns correct values
	pathAttr, err := p.Attr("path")
	if err != nil {
		t.Errorf("Attr(path) error = %v", err)
	}
	if pathAttr.(starlark.String).GoString() != "/foo/bar/baz.txt" {
		t.Errorf("Attr(path) = %v, want /foo/bar/baz.txt", pathAttr)
	}

	nameAttr, err := p.Attr("name")
	if err != nil {
		t.Errorf("Attr(name) error = %v", err)
	}
	if nameAttr.(starlark.String).GoString() != "baz.txt" {
		t.Errorf("Attr(name) = %v, want baz.txt", nameAttr)
	}

	// Test unknown attr returns nil
	unknownAttr, err := p.Attr("unknown")
	if err != nil {
		t.Errorf("Attr(unknown) error = %v", err)
	}
	if unknownAttr != nil {
		t.Errorf("Attr(unknown) = %v, want nil", unknownAttr)
	}
}

// TestChangeStarlarkValue tests that Change implements starlark.Value correctly.
func TestChangeStarlarkValue(t *testing.T) {
	author := authoring.NewAuthor("John Doe", "john@example.com")
	dt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	c := NewChange("abc123", author, "Fix bug", dt, []string{"main.go", "util.go"})

	// Test String()
	if got := c.String(); got != "change<abc123>" {
		t.Errorf("String() = %q, want %q", got, "change<abc123>")
	}

	// Test Type()
	if got := c.Type(); got != "change" {
		t.Errorf("Type() = %q, want %q", got, "change")
	}

	// Test Truth()
	if got := c.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	// Test Hash()
	hash, err := c.Hash()
	if err != nil {
		t.Errorf("Hash() error = %v", err)
	}
	if hash == 0 {
		t.Error("Hash() should not be 0")
	}
}

// TestChangeMethods tests Change methods.
func TestChangeMethods(t *testing.T) {
	author := authoring.NewAuthor("John Doe", "john@example.com")
	dt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	c := NewChange("abc123", author, "Fix bug", dt, []string{"main.go", "util.go"})

	if c.Ref() != "abc123" {
		t.Errorf("Ref() = %q, want %q", c.Ref(), "abc123")
	}

	if c.Author().Name() != "John Doe" {
		t.Errorf("Author().Name() = %q, want %q", c.Author().Name(), "John Doe")
	}

	if c.Message().Text() != "Fix bug" {
		t.Errorf("Message().Text() = %q, want %q", c.Message().Text(), "Fix bug")
	}

	if !c.DateTime().Equal(dt) {
		t.Errorf("DateTime() = %v, want %v", c.DateTime(), dt)
	}

	files := c.Files()
	if len(files) != 2 || files[0] != "main.go" {
		t.Errorf("Files() = %v, want [main.go, util.go]", files)
	}
}

// TestChangeHasAttrs tests that Change implements starlark.HasAttrs.
func TestChangeHasAttrs(t *testing.T) {
	author := authoring.NewAuthor("John Doe", "john@example.com")
	dt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	c := NewChange("abc123", author, "Fix bug", dt, []string{"main.go"})

	// Test AttrNames
	names := c.AttrNames()
	expectedNames := []string{"author", "date_time", "files", "message", "ref"}
	if !slices.Equal(names, expectedNames) {
		t.Errorf("AttrNames() = %v, want %v", names, expectedNames)
	}

	// Test ref attr
	refAttr, err := c.Attr("ref")
	if err != nil {
		t.Errorf("Attr(ref) error = %v", err)
	}
	if refAttr.(starlark.String).GoString() != "abc123" {
		t.Errorf("Attr(ref) = %v, want abc123", refAttr)
	}

	// Test author attr
	authorAttr, err := c.Attr("author")
	if err != nil {
		t.Errorf("Attr(author) error = %v", err)
	}
	if _, ok := authorAttr.(*authoring.Author); !ok {
		t.Errorf("Attr(author) is not *authoring.Author")
	}
}

// TestChangeMessageStarlarkValue tests that ChangeMessage implements starlark.Value correctly.
func TestChangeMessageStarlarkValue(t *testing.T) {
	cm := NewChangeMessage("Fix bug\n\nBUG=123\nREVIEWER=alice")

	// Test String()
	if got := cm.String(); got != "Fix bug\n\nBUG=123\nREVIEWER=alice" {
		t.Errorf("String() = %q", got)
	}

	// Test Type()
	if got := cm.Type(); got != "change_message" {
		t.Errorf("Type() = %q, want %q", got, "change_message")
	}

	// Test Truth()
	if got := cm.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	emptyCM := NewChangeMessage("")
	if got := emptyCM.Truth(); got != starlark.False {
		t.Errorf("empty message Truth() = %v, want False", got)
	}

	// Test Hash()
	hash, err := cm.Hash()
	if err != nil {
		t.Errorf("Hash() error = %v", err)
	}
	if hash == 0 {
		t.Error("Hash() should not be 0")
	}
}

// TestChangeMessageMethods tests ChangeMessage methods.
func TestChangeMessageMethods(t *testing.T) {
	cm := NewChangeMessage("Fix bug\n\nThis fixes the issue.\n\nBUG=123\nBUG=456\nREVIEWER: alice")

	// Test Text()
	if !contains(cm.Text(), "Fix bug") {
		t.Errorf("Text() doesn't contain expected content")
	}

	// Test FirstLine()
	if cm.FirstLine() != "Fix bug" {
		t.Errorf("FirstLine() = %q, want %q", cm.FirstLine(), "Fix bug")
	}

	// Test GetLabel()
	if cm.GetLabel("BUG") != "123" {
		t.Errorf("GetLabel(BUG) = %q, want %q", cm.GetLabel("BUG"), "123")
	}
	if cm.GetLabel("REVIEWER") != "alice" {
		t.Errorf("GetLabel(REVIEWER) = %q, want %q", cm.GetLabel("REVIEWER"), "alice")
	}
	if cm.GetLabel("NONEXISTENT") != "" {
		t.Errorf("GetLabel(NONEXISTENT) = %q, want empty", cm.GetLabel("NONEXISTENT"))
	}

	// Test GetLabelAll()
	bugLabels := cm.GetLabelAll("BUG")
	if len(bugLabels) != 2 || bugLabels[0] != "123" || bugLabels[1] != "456" {
		t.Errorf("GetLabelAll(BUG) = %v, want [123, 456]", bugLabels)
	}

	// Test Labels()
	labels := cm.Labels()
	if len(labels["BUG"]) != 2 {
		t.Errorf("Labels()[BUG] = %v, want 2 items", labels["BUG"])
	}
}

// TestChangeMessageHasAttrs tests that ChangeMessage implements starlark.HasAttrs.
func TestChangeMessageHasAttrs(t *testing.T) {
	cm := NewChangeMessage("Fix bug\n\nBUG=123")

	// Test AttrNames
	names := cm.AttrNames()
	expectedNames := []string{"first_line", "get_label", "get_label_all", "labels", "text"}
	if !slices.Equal(names, expectedNames) {
		t.Errorf("AttrNames() = %v, want %v", names, expectedNames)
	}

	// Test text attr
	textAttr, err := cm.Attr("text")
	if err != nil {
		t.Errorf("Attr(text) error = %v", err)
	}
	if !contains(textAttr.(starlark.String).GoString(), "Fix bug") {
		t.Errorf("Attr(text) doesn't contain expected content")
	}

	// Test first_line attr
	firstLineAttr, err := cm.Attr("first_line")
	if err != nil {
		t.Errorf("Attr(first_line) error = %v", err)
	}
	if firstLineAttr.(starlark.String).GoString() != "Fix bug" {
		t.Errorf("Attr(first_line) = %v, want Fix bug", firstLineAttr)
	}

	// Test labels attr
	labelsAttr, err := cm.Attr("labels")
	if err != nil {
		t.Errorf("Attr(labels) error = %v", err)
	}
	if _, ok := labelsAttr.(*starlark.Dict); !ok {
		t.Error("Attr(labels) is not *starlark.Dict")
	}
}

// TestOriginRefStarlarkValue tests that OriginRef implements starlark.Value correctly.
func TestOriginRefStarlarkValue(t *testing.T) {
	o := NewOriginRef("abc123", "https://github.com/example/repo")

	// Test String()
	if got := o.String(); got != "origin_ref<abc123>" {
		t.Errorf("String() = %q, want %q", got, "origin_ref<abc123>")
	}

	// Test Type()
	if got := o.Type(); got != "origin_ref" {
		t.Errorf("Type() = %q, want %q", got, "origin_ref")
	}

	// Test Truth()
	if got := o.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	emptyRef := NewOriginRef("", "")
	if got := emptyRef.Truth(); got != starlark.False {
		t.Errorf("empty ref Truth() = %v, want False", got)
	}

	// Test Hash()
	hash, err := o.Hash()
	if err != nil {
		t.Errorf("Hash() error = %v", err)
	}
	if hash == 0 {
		t.Error("Hash() should not be 0")
	}
}

// TestOriginRefMethods tests OriginRef methods.
func TestOriginRefMethods(t *testing.T) {
	o := NewOriginRef("abc123", "https://github.com/example/repo")

	if o.Ref() != "abc123" {
		t.Errorf("Ref() = %q, want %q", o.Ref(), "abc123")
	}

	if o.URL() != "https://github.com/example/repo" {
		t.Errorf("URL() = %q, want %q", o.URL(), "https://github.com/example/repo")
	}
}

// TestOriginRefHasAttrs tests that OriginRef implements starlark.HasAttrs.
func TestOriginRefHasAttrs(t *testing.T) {
	o := NewOriginRef("abc123", "https://github.com/example/repo")

	// Test AttrNames
	names := o.AttrNames()
	expectedNames := []string{"ref", "url"}
	if !slices.Equal(names, expectedNames) {
		t.Errorf("AttrNames() = %v, want %v", names, expectedNames)
	}

	// Test ref attr
	refAttr, err := o.Attr("ref")
	if err != nil {
		t.Errorf("Attr(ref) error = %v", err)
	}
	if refAttr.(starlark.String).GoString() != "abc123" {
		t.Errorf("Attr(ref) = %v, want abc123", refAttr)
	}

	// Test url attr
	urlAttr, err := o.Attr("url")
	if err != nil {
		t.Errorf("Attr(url) error = %v", err)
	}
	if urlAttr.(starlark.String).GoString() != "https://github.com/example/repo" {
		t.Errorf("Attr(url) = %v", urlAttr)
	}
}

// TestDestinationRefStarlarkValue tests that DestinationRef implements starlark.Value correctly.
func TestDestinationRefStarlarkValue(t *testing.T) {
	d := NewDestinationRef("def456", "https://github.com/dest/repo")

	// Test String()
	if got := d.String(); got != "destination_ref<def456>" {
		t.Errorf("String() = %q, want %q", got, "destination_ref<def456>")
	}

	// Test Type()
	if got := d.Type(); got != "destination_ref" {
		t.Errorf("Type() = %q, want %q", got, "destination_ref")
	}

	// Test Truth()
	if got := d.Truth(); got != starlark.True {
		t.Errorf("Truth() = %v, want True", got)
	}

	emptyRef := NewDestinationRef("", "")
	if got := emptyRef.Truth(); got != starlark.False {
		t.Errorf("empty ref Truth() = %v, want False", got)
	}

	// Test Hash()
	hash, err := d.Hash()
	if err != nil {
		t.Errorf("Hash() error = %v", err)
	}
	if hash == 0 {
		t.Error("Hash() should not be 0")
	}
}

// TestDestinationRefMethods tests DestinationRef methods.
func TestDestinationRefMethods(t *testing.T) {
	d := NewDestinationRef("def456", "https://github.com/dest/repo")

	if d.Ref() != "def456" {
		t.Errorf("Ref() = %q, want %q", d.Ref(), "def456")
	}

	if d.URL() != "https://github.com/dest/repo" {
		t.Errorf("URL() = %q, want %q", d.URL(), "https://github.com/dest/repo")
	}
}

// TestDestinationRefHasAttrs tests that DestinationRef implements starlark.HasAttrs.
func TestDestinationRefHasAttrs(t *testing.T) {
	d := NewDestinationRef("def456", "https://github.com/dest/repo")

	// Test AttrNames
	names := d.AttrNames()
	expectedNames := []string{"ref", "url"}
	if !slices.Equal(names, expectedNames) {
		t.Errorf("AttrNames() = %v, want %v", names, expectedNames)
	}

	// Test ref attr
	refAttr, err := d.Attr("ref")
	if err != nil {
		t.Errorf("Attr(ref) error = %v", err)
	}
	if refAttr.(starlark.String).GoString() != "def456" {
		t.Errorf("Attr(ref) = %v, want def456", refAttr)
	}

	// Test url attr
	urlAttr, err := d.Attr("url")
	if err != nil {
		t.Errorf("Attr(url) error = %v", err)
	}
	if urlAttr.(starlark.String).GoString() != "https://github.com/dest/repo" {
		t.Errorf("Attr(url) = %v", urlAttr)
	}
}

// TestInterfaceCompliance tests that all types implement the expected interfaces.
func TestInterfaceCompliance(t *testing.T) {
	// These assignments will fail at compile time if interfaces are not implemented
	var _ starlark.Value = (*Path)(nil)
	var _ starlark.HasAttrs = (*Path)(nil)

	var _ starlark.Value = (*Change)(nil)
	var _ starlark.HasAttrs = (*Change)(nil)

	var _ starlark.Value = (*ChangeMessage)(nil)
	var _ starlark.HasAttrs = (*ChangeMessage)(nil)

	var _ starlark.Value = (*OriginRef)(nil)
	var _ starlark.HasAttrs = (*OriginRef)(nil)

	var _ starlark.Value = (*DestinationRef)(nil)
	var _ starlark.HasAttrs = (*DestinationRef)(nil)
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Package metadata provides metadata transformation types.
//
// Reference: https://github.com/google/copybara/tree/master/java/com/google/copybara/transform/metadata
package metadata

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"go.starlark.net/starlark"

	"github.com/albertocavalcante/starlark-go-copybara/authoring"
	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

// SquashNotes is a transformation that squashes commit notes.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/MetadataSquashNotes.java
type SquashNotes struct {
	prefix          string
	max             int
	compact         bool
	showRef         bool
	showAuthor      bool
	showDescription bool
	oldestFirst     bool
	useMerge        bool
}

func (s *SquashNotes) String() string { return "metadata.squash_notes()" }
func (s *SquashNotes) Type() string   { return "squash_notes" }
func (s *SquashNotes) Freeze()        {}
func (s *SquashNotes) Truth() starlark.Bool {
	return starlark.True
}
func (s *SquashNotes) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: squash_notes")
}

// Apply implements Transformation.
func (s *SquashNotes) Apply(ctx *transform.Context) error {
	var sb strings.Builder

	// Resolve prefix template
	prefix := s.resolveTemplate(s.prefix, ctx)
	sb.WriteString(prefix)

	if s.max == 0 {
		ctx.Message = sb.String()
		return nil
	}

	changes := ctx.Changes.Current
	if s.oldestFirst {
		changes = slices.Clone(changes)
		slices.Reverse(changes)
	}

	// Filter out merges if needed
	if !s.useMerge {
		var filtered []*transform.Change
		for _, c := range changes {
			if !c.IsMerge {
				filtered = append(filtered, c)
			}
		}
		changes = filtered
	}

	counter := 0
	for i, c := range changes {
		if counter >= s.max {
			break
		}

		author := c.MappedAuthor
		if author == "" {
			author = c.Author
		}

		if s.compact {
			sb.WriteString("  - ")
			var parts []string
			if s.showRef {
				parts = append(parts, c.Ref)
			}
			if s.showDescription {
				parts = append(parts, cutIfLong(c.FirstLineMessage()))
			}
			if s.showAuthor {
				parts = append(parts, "by "+author)
			}
			sb.WriteString(strings.Join(parts, " "))
			sb.WriteString("\n")
		} else {
			sb.WriteString("--\n")
			var parts []string
			if s.showRef {
				parts = append(parts, c.Ref)
			} else {
				parts = append(parts, fmt.Sprintf("Change %d of %d", i+1, len(changes)))
			}
			if s.showAuthor {
				parts = append(parts, "by "+author)
			}
			sb.WriteString(strings.Join(parts, " "))
			if s.showDescription {
				sb.WriteString(":\n\n")
				sb.WriteString(c.Message)
			}
			sb.WriteString("\n")
		}
		counter++
	}

	if len(changes) > s.max {
		sb.WriteString(fmt.Sprintf("  (And %d more changes)\n", len(changes)-s.max))
	}

	ctx.Message = sb.String()
	return nil
}

func (s *SquashNotes) resolveTemplate(template string, ctx *transform.Context) string {
	// Simple template resolution for ${LABEL_NAME} patterns
	re := regexp.MustCompile(`\$\{([A-Za-z][A-Za-z0-9_-]*)\}`)
	return re.ReplaceAllStringFunc(template, func(match string) string {
		label := match[2 : len(match)-1] // Extract label name from ${...}
		if value := ctx.GetLabel(label); value != "" {
			return value
		}
		return match
	})
}

func cutIfLong(msg string) string {
	if len(msg) < 60 {
		return msg
	}
	return msg[:57] + "..."
}

// Reverse implements Transformation.
func (s *SquashNotes) Reverse() transform.Transformation {
	return transform.NewNoopTransformation(s)
}

// Describe implements Transformation.
func (s *SquashNotes) Describe() string {
	return "squash_notes"
}

// SaveAuthor is a transformation that saves the original author.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/SaveOriginalAuthor.java
type SaveAuthor struct {
	label     string
	separator string
}

func (s *SaveAuthor) String() string {
	return fmt.Sprintf("metadata.save_author(label = %q)", s.label)
}
func (s *SaveAuthor) Type() string          { return "save_author" }
func (s *SaveAuthor) Freeze()               {}
func (s *SaveAuthor) Truth() starlark.Bool  { return starlark.True }
func (s *SaveAuthor) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: save_author") }

// Apply implements Transformation.
func (s *SaveAuthor) Apply(ctx *transform.Context) error {
	if ctx.Author != "" {
		ctx.AddLabel(s.label, ctx.Author, s.separator)
	}
	return nil
}

// Reverse implements Transformation.
func (s *SaveAuthor) Reverse() transform.Transformation {
	return &RestoreAuthor{
		label:            s.label,
		separator:        s.separator,
		searchAllChanges: false,
	}
}

// Describe implements Transformation.
func (s *SaveAuthor) Describe() string {
	return fmt.Sprintf("Saving author as label '%s'", s.label)
}

// RestoreAuthor is a transformation that restores the original author from a label.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/RestoreOriginalAuthor.java
type RestoreAuthor struct {
	label            string
	separator        string
	searchAllChanges bool
}

func (r *RestoreAuthor) String() string {
	return fmt.Sprintf("metadata.restore_author(label = %q)", r.label)
}
func (r *RestoreAuthor) Type() string         { return "restore_author" }
func (r *RestoreAuthor) Freeze()              {}
func (r *RestoreAuthor) Truth() starlark.Bool { return starlark.True }
func (r *RestoreAuthor) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: restore_author")
}

// Apply implements Transformation.
func (r *RestoreAuthor) Apply(ctx *transform.Context) error {
	var authorStr string

	// Search in changes
	for _, change := range ctx.Changes.Current {
		if change.Labels != nil {
			if values, ok := change.Labels[r.label]; ok && len(values) > 0 {
				authorStr = values[len(values)-1] // Last value wins
			}
		}
		if !r.searchAllChanges {
			break
		}
	}

	// Also check the message
	if authorStr == "" {
		authorStr = ctx.GetLabel(r.label)
	}

	if authorStr != "" {
		author, err := authoring.ParseAuthor(authorStr)
		if err == nil {
			ctx.Author = author.String()
			ctx.RemoveLabel(r.label)
		}
		// If parsing fails, we silently continue without updating the author
	}

	return nil
}

// Reverse implements Transformation.
func (r *RestoreAuthor) Reverse() transform.Transformation {
	return &SaveAuthor{
		label:     r.label,
		separator: r.separator,
	}
}

// Describe implements Transformation.
func (r *RestoreAuthor) Describe() string {
	return "Restoring original author"
}

// ReplaceMessage is a transformation that replaces the commit message.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/TemplateMessage.java
type ReplaceMessage struct {
	message             string
	ignoreLabelNotFound bool
}

func (r *ReplaceMessage) String() string {
	return fmt.Sprintf("metadata.replace_message(%q)", r.message)
}
func (r *ReplaceMessage) Type() string         { return "replace_message" }
func (r *ReplaceMessage) Freeze()              {}
func (r *ReplaceMessage) Truth() starlark.Bool { return starlark.True }
func (r *ReplaceMessage) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: replace_message")
}

// Apply implements Transformation.
func (r *ReplaceMessage) Apply(ctx *transform.Context) error {
	resolved, err := resolveTemplateLabels(r.message, ctx, r.ignoreLabelNotFound)
	if err != nil {
		return err
	}
	if resolved != "" || !r.ignoreLabelNotFound {
		ctx.Message = resolved
	}
	return nil
}

// Reverse implements Transformation.
func (r *ReplaceMessage) Reverse() transform.Transformation {
	return transform.NewNoopTransformation(r)
}

// Describe implements Transformation.
func (r *ReplaceMessage) Describe() string {
	return "Replacing commit message"
}

// ExposeLabel is a transformation that exposes a label from the commit message.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/ExposeLabelInMessage.java
type ExposeLabel struct {
	name                string
	newName             string
	separator           string
	ignoreLabelNotFound bool
	all                 bool
	concatSeparator     string
	hasConcatSeparator  bool
}

func (e *ExposeLabel) String() string {
	return fmt.Sprintf("metadata.expose_label(%q)", e.name)
}
func (e *ExposeLabel) Type() string          { return "expose_label" }
func (e *ExposeLabel) Freeze()               {}
func (e *ExposeLabel) Truth() starlark.Bool  { return starlark.True }
func (e *ExposeLabel) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: expose_label") }

// Apply implements Transformation.
func (e *ExposeLabel) Apply(ctx *transform.Context) error {
	if e.all {
		return e.exposeAllLabels(ctx)
	}

	value := ctx.GetLabel(e.name)
	if value == "" {
		if e.ignoreLabelNotFound {
			return nil
		}
		return fmt.Errorf("cannot find label %s", e.name)
	}

	// If the label name is the same, remove it first
	if e.name == e.newName {
		ctx.RemoveLabelWithValue(e.name, value)
	}

	ctx.AddLabel(e.newName, value, e.separator)
	return nil
}

func (e *ExposeLabel) exposeAllLabels(ctx *transform.Context) error {
	values := ctx.GetAllLabels(e.name)

	if len(values) == 0 {
		if e.ignoreLabelNotFound {
			return nil
		}
		return fmt.Errorf("cannot find label %s", e.name)
	}

	// If the label name is the same, remove all instances first
	if e.name == e.newName {
		ctx.RemoveLabel(e.name)
	}

	if e.hasConcatSeparator {
		ctx.AddLabel(e.newName, strings.Join(values, e.concatSeparator), e.separator)
	} else {
		for _, value := range values {
			ctx.AddLabel(e.newName, value, e.separator)
		}
	}

	return nil
}

// Reverse implements Transformation.
func (e *ExposeLabel) Reverse() transform.Transformation {
	return transform.NewNoopTransformation(e)
}

// Describe implements Transformation.
func (e *ExposeLabel) Describe() string {
	return fmt.Sprintf("Exposing label %s as %s", e.name, e.newName)
}

// AddHeader is a transformation that adds a header to the commit message.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/TemplateMessage.java
type AddHeader struct {
	text                string
	ignoreLabelNotFound bool
	newLine             bool
}

func (a *AddHeader) String() string {
	return fmt.Sprintf("metadata.add_header(%q)", a.text)
}
func (a *AddHeader) Type() string          { return "add_header" }
func (a *AddHeader) Freeze()               {}
func (a *AddHeader) Truth() starlark.Bool  { return starlark.True }
func (a *AddHeader) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: add_header") }

// Apply implements Transformation.
func (a *AddHeader) Apply(ctx *transform.Context) error {
	resolved, err := resolveTemplateLabels(a.text, ctx, a.ignoreLabelNotFound)
	if err != nil {
		return err
	}
	if resolved == "" && a.ignoreLabelNotFound {
		return nil
	}

	var sb strings.Builder
	sb.WriteString(resolved)
	if a.newLine {
		sb.WriteString("\n")
	}
	sb.WriteString(ctx.Message)

	ctx.Message = sb.String()
	return nil
}

// Reverse implements Transformation.
func (a *AddHeader) Reverse() transform.Transformation {
	return transform.NewNoopTransformation(a)
}

// Describe implements Transformation.
func (a *AddHeader) Describe() string {
	return "Adding header to commit message"
}

// Scrubber is a transformation that scrubs sensitive content from commit messages.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/Scrubber.java
type Scrubber struct {
	regexes       []*regexp.Regexp
	replacement   string
	msgIfNoMatch  string
	failIfNoMatch bool
}

func (s *Scrubber) String() string        { return "metadata.scrubber()" }
func (s *Scrubber) Type() string          { return "scrubber" }
func (s *Scrubber) Freeze()               {}
func (s *Scrubber) Truth() starlark.Bool  { return starlark.True }
func (s *Scrubber) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: scrubber") }

// Apply implements Transformation.
func (s *Scrubber) Apply(ctx *transform.Context) error {
	message := ctx.Message
	matched := false

	for _, re := range s.regexes {
		newMessage := re.ReplaceAllString(message, s.replacement)
		if newMessage != message {
			matched = true
			message = newMessage
		}
	}

	if matched {
		ctx.Message = message
		return nil
	}

	if s.failIfNoMatch {
		return fmt.Errorf("scrubber regex didn't match for description: %s", ctx.Message)
	}

	if s.msgIfNoMatch != "" {
		ctx.Message = s.msgIfNoMatch
	}

	return nil
}

// Reverse implements Transformation.
func (s *Scrubber) Reverse() transform.Transformation {
	return transform.NewNoopTransformation(s)
}

// Describe implements Transformation.
func (s *Scrubber) Describe() string {
	return "Description scrubber"
}

// MapAuthor is a transformation that maps author identities.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/transform/metadata/MapAuthor.java
type MapAuthor struct {
	// authorToAuthor maps full "Name <email>" to "Name <email>"
	authorToAuthor map[string]string
	// mailToAuthor maps email only to Author
	mailToAuthor map[string]*authoring.Author
	// nameToAuthor maps name only to Author
	nameToAuthor map[string]*authoring.Author

	reversible            bool
	noopReverse           bool
	failIfNotFound        bool
	reverseFailIfNotFound bool
	mapAll                bool
}

func (m *MapAuthor) String() string        { return "metadata.map_author()" }
func (m *MapAuthor) Type() string          { return "map_author" }
func (m *MapAuthor) Freeze()               {}
func (m *MapAuthor) Truth() starlark.Bool  { return starlark.True }
func (m *MapAuthor) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: map_author") }

// NewMapAuthor creates a new MapAuthor transformation from an author mapping.
func NewMapAuthor(authors map[string]string, reversible, noopReverse, failIfNotFound, reverseFailIfNotFound, mapAll bool) (*MapAuthor, error) {
	authorToAuthor := make(map[string]string)
	mailToAuthor := make(map[string]*authoring.Author)
	nameToAuthor := make(map[string]*authoring.Author)

	for key, value := range authors {
		toAuthor, err := authoring.ParseAuthor(value)
		if err != nil {
			return nil, fmt.Errorf("invalid author value %q: %w", value, err)
		}

		// Try to parse the key as a full author
		fromAuthor, err := authoring.ParseAuthor(key)
		if err == nil {
			authorToAuthor[fromAuthor.String()] = toAuthor.String()
		} else if strings.Contains(key, "@") {
			mailToAuthor[key] = toAuthor
		} else {
			nameToAuthor[key] = toAuthor
		}
	}

	return &MapAuthor{
		authorToAuthor:        authorToAuthor,
		mailToAuthor:          mailToAuthor,
		nameToAuthor:          nameToAuthor,
		reversible:            reversible,
		noopReverse:           noopReverse,
		failIfNotFound:        failIfNotFound,
		reverseFailIfNotFound: reverseFailIfNotFound,
		mapAll:                mapAll,
	}, nil
}

// Apply implements Transformation.
func (m *MapAuthor) Apply(ctx *transform.Context) error {
	author, err := m.getMappedAuthor(ctx.Author)
	if err != nil {
		return err
	}
	ctx.Author = author

	if m.mapAll {
		for _, change := range ctx.Changes.Current {
			mapped, err := m.getMappedAuthor(change.Author)
			if err != nil {
				return err
			}
			change.MappedAuthor = mapped
		}
	}

	return nil
}

func (m *MapAuthor) getMappedAuthor(authorStr string) (string, error) {
	// Try exact author match
	if newAuthor, ok := m.authorToAuthor[authorStr]; ok {
		return newAuthor, nil
	}

	// Parse the author
	author, err := authoring.ParseAuthor(authorStr)
	if err != nil {
		if m.failIfNotFound {
			return "", fmt.Errorf("cannot find a mapping for author '%s'", authorStr)
		}
		return authorStr, nil
	}

	// Try email match
	if mapped, ok := m.mailToAuthor[author.Email()]; ok {
		return mapped.String(), nil
	}

	// Try name match
	if mapped, ok := m.nameToAuthor[author.Name()]; ok {
		return mapped.String(), nil
	}

	if m.failIfNotFound {
		return "", fmt.Errorf("cannot find a mapping for author '%s'", authorStr)
	}
	return authorStr, nil
}

// Reverse implements Transformation.
func (m *MapAuthor) Reverse() transform.Transformation {
	if m.noopReverse {
		return transform.NewNoopTransformation(m)
	}

	if !m.reversible {
		return transform.NewErrorTransformation(fmt.Errorf("author mapping doesn't have reversible enabled"), m)
	}

	if len(m.mailToAuthor) > 0 {
		return transform.NewErrorTransformation(fmt.Errorf("author mapping is not reversible because it contains mail -> author mappings"), m)
	}

	if len(m.nameToAuthor) > 0 {
		return transform.NewErrorTransformation(fmt.Errorf("author mapping is not reversible because it contains name -> author mappings"), m)
	}

	// Build reverse mapping
	reverse := make(map[string]string)
	for k, v := range m.authorToAuthor {
		if _, exists := reverse[v]; exists {
			return transform.NewErrorTransformation(fmt.Errorf("non-reversible author map: duplicate target author"), m)
		}
		reverse[v] = k
	}

	return &MapAuthor{
		authorToAuthor:        reverse,
		mailToAuthor:          make(map[string]*authoring.Author),
		nameToAuthor:          make(map[string]*authoring.Author),
		reversible:            m.reversible,
		noopReverse:           m.noopReverse,
		failIfNotFound:        m.reverseFailIfNotFound,
		reverseFailIfNotFound: m.failIfNotFound,
		mapAll:                m.mapAll,
	}
}

// Describe implements Transformation.
func (m *MapAuthor) Describe() string {
	return "Mapping authors"
}

// resolveTemplateLabels resolves ${LABEL_NAME} patterns in the template.
func resolveTemplateLabels(template string, ctx *transform.Context, ignoreLabelNotFound bool) (string, error) {
	re := regexp.MustCompile(`\$\{([A-Za-z][A-Za-z0-9_-]*)\}`)

	var lastErr error
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		label := match[2 : len(match)-1] // Extract label name from ${...}
		value := ctx.GetLabel(label)
		if value == "" {
			if !ignoreLabelNotFound {
				lastErr = fmt.Errorf("cannot find label '%s' in message or changes", label)
			}
			return match
		}
		return value
	})

	if lastErr != nil {
		return "", lastErr
	}

	return result, nil
}

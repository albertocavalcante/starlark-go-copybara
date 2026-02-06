package types

import (
	"regexp"
	"strings"

	"go.starlark.net/starlark"
)

// Ensure ChangeMessage implements required interfaces.
var (
	_ starlark.Value    = (*ChangeMessage)(nil)
	_ starlark.HasAttrs = (*ChangeMessage)(nil)
)

// labelPattern matches labels in the format "LABEL=value" or "LABEL: value".
// Labels are typically at the end of commit messages, one per line.
var labelPattern = regexp.MustCompile(`(?m)^([A-Z][A-Z0-9_-]*)\s*[:=]\s*(.+)$`)

// ChangeMessage wraps a commit message with label support.
//
// Reference: https://github.com/google/copybara/blob/master/java/com/google/copybara/ChangeMessage.java
type ChangeMessage struct {
	text   string
	labels map[string][]string
}

// NewChangeMessage creates a new ChangeMessage from raw text.
func NewChangeMessage(text string) *ChangeMessage {
	cm := &ChangeMessage{
		text:   text,
		labels: make(map[string][]string),
	}
	cm.parseLabels()
	return cm
}

// parseLabels extracts labels from the message text.
func (cm *ChangeMessage) parseLabels() {
	matches := labelPattern.FindAllStringSubmatch(cm.text, -1)
	for _, match := range matches {
		label := match[1]
		value := strings.TrimSpace(match[2])
		cm.labels[label] = append(cm.labels[label], value)
	}
}

// String implements starlark.Value.
func (cm *ChangeMessage) String() string {
	return cm.text
}

// Type implements starlark.Value.
func (cm *ChangeMessage) Type() string {
	return "change_message"
}

// Freeze implements starlark.Value.
func (cm *ChangeMessage) Freeze() {}

// Truth implements starlark.Value.
func (cm *ChangeMessage) Truth() starlark.Bool {
	return starlark.Bool(cm.text != "")
}

// Hash implements starlark.Value.
func (cm *ChangeMessage) Hash() (uint32, error) {
	return starlark.String(cm.text).Hash()
}

// Text returns the full message text.
func (cm *ChangeMessage) Text() string {
	return cm.text
}

// FirstLine returns the first line of the message (the subject).
func (cm *ChangeMessage) FirstLine() string {
	if idx := strings.IndexByte(cm.text, '\n'); idx != -1 {
		return cm.text[:idx]
	}
	return cm.text
}

// Labels returns all parsed labels as a map.
func (cm *ChangeMessage) Labels() map[string][]string {
	// Return a copy to prevent modification
	result := make(map[string][]string, len(cm.labels))
	for k, v := range cm.labels {
		result[k] = append([]string(nil), v...)
	}
	return result
}

// GetLabel returns the first value for the given label name, or empty string if not found.
func (cm *ChangeMessage) GetLabel(name string) string {
	if values, ok := cm.labels[name]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// GetLabelAll returns all values for the given label name.
func (cm *ChangeMessage) GetLabelAll(name string) []string {
	if values, ok := cm.labels[name]; ok {
		return append([]string(nil), values...)
	}
	return nil
}

// Attr implements starlark.HasAttrs.
func (cm *ChangeMessage) Attr(name string) (starlark.Value, error) {
	switch name {
	case "text":
		return starlark.String(cm.text), nil
	case "first_line":
		return starlark.String(cm.FirstLine()), nil
	case "labels":
		return cm.labelsToDict(), nil
	case "get_label":
		return starlark.NewBuiltin("change_message.get_label", cm.getLabelBuiltin), nil
	case "get_label_all":
		return starlark.NewBuiltin("change_message.get_label_all", cm.getLabelAllBuiltin), nil
	default:
		return nil, nil
	}
}

// AttrNames implements starlark.HasAttrs.
func (cm *ChangeMessage) AttrNames() []string {
	return []string{"first_line", "get_label", "get_label_all", "labels", "text"}
}

// labelsToDict converts labels to a starlark.Dict.
func (cm *ChangeMessage) labelsToDict() *starlark.Dict {
	dict := starlark.NewDict(len(cm.labels))
	for k, v := range cm.labels {
		_ = dict.SetKey(starlark.String(k), StringSliceToList(v))
	}
	return dict
}

// getLabelBuiltin is the Starlark builtin for get_label.
func (cm *ChangeMessage) getLabelBuiltin(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	return starlark.String(cm.GetLabel(name)), nil
}

// getLabelAllBuiltin is the Starlark builtin for get_label_all.
func (cm *ChangeMessage) getLabelAllBuiltin(_ *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "name", &name); err != nil {
		return nil, err
	}
	values := cm.GetLabelAll(name)
	if values == nil {
		return starlark.NewList(nil), nil
	}
	return StringSliceToList(values), nil
}

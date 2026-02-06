package transform_test

import (
	"testing"

	"github.com/albertocavalcante/starlark-go-copybara/transform"
)

func TestNewContext(t *testing.T) {
	ctx := transform.NewContext("/tmp/work")

	if ctx.WorkDir != "/tmp/work" {
		t.Errorf("expected WorkDir '/tmp/work', got %q", ctx.WorkDir)
	}

	if ctx.Labels == nil {
		t.Error("expected Labels to be initialized")
	}

	if ctx.Changes == nil {
		t.Error("expected Changes to be initialized")
	}
}

func TestChangeFirstLineMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "single line",
			message:  "Single line message",
			expected: "Single line message",
		},
		{
			name:     "multi line",
			message:  "First line\nSecond line\nThird line",
			expected: "First line",
		},
		{
			name:     "empty message",
			message:  "",
			expected: "",
		},
		{
			name:     "only newline",
			message:  "\n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			change := &transform.Change{Message: tt.message}
			result := change.FirstLineMessage()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetLabel(t *testing.T) {
	ctx := transform.NewContext("/tmp")

	// Test pre-populated label
	ctx.Labels["PRESET"] = []string{"preset_value"}
	if got := ctx.GetLabel("PRESET"); got != "preset_value" {
		t.Errorf("expected 'preset_value', got %q", got)
	}

	// Test label from message
	ctx.Message = "Some message\nLABEL_NAME=label_value\nMore text"
	if got := ctx.GetLabel("LABEL_NAME"); got != "label_value" {
		t.Errorf("expected 'label_value', got %q", got)
	}

	// Test label with colon separator
	ctx.Message = "Some message\nOTHER_LABEL: colon_value\n"
	if got := ctx.GetLabel("OTHER_LABEL"); got != "colon_value" {
		t.Errorf("expected 'colon_value', got %q", got)
	}

	// Test label from changes
	ctx.Message = ""
	ctx.Labels = make(map[string][]string)
	ctx.Changes.Current = []*transform.Change{
		{
			Labels: map[string][]string{
				"CHANGE_LABEL": {"change_value"},
			},
		},
	}
	if got := ctx.GetLabel("CHANGE_LABEL"); got != "change_value" {
		t.Errorf("expected 'change_value', got %q", got)
	}

	// Test non-existent label
	if got := ctx.GetLabel("NON_EXISTENT"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetAllLabels(t *testing.T) {
	ctx := transform.NewContext("/tmp")

	// Set up multiple sources of labels
	ctx.Labels["LABEL"] = []string{"preset1", "preset2"}
	ctx.Message = "Message\nLABEL=msg_value\n"
	ctx.Changes.Current = []*transform.Change{
		{
			Labels: map[string][]string{
				"LABEL": {"change1", "change2"},
			},
		},
	}

	values := ctx.GetAllLabels("LABEL")

	// Should contain unique values from all sources
	expected := map[string]bool{
		"preset1":   true,
		"preset2":   true,
		"msg_value": true,
		"change1":   true,
		"change2":   true,
	}

	if len(values) != len(expected) {
		t.Errorf("expected %d values, got %d: %v", len(expected), len(values), values)
	}

	for _, v := range values {
		if !expected[v] {
			t.Errorf("unexpected value %q in results", v)
		}
	}
}

func TestAddLabel(t *testing.T) {
	ctx := transform.NewContext("/tmp")
	ctx.Message = "Original message"

	ctx.AddLabel("NEW_LABEL", "new_value", "=")

	// Check labels map
	if values, ok := ctx.Labels["NEW_LABEL"]; !ok || len(values) == 0 || values[0] != "new_value" {
		t.Errorf("expected label to be added to Labels map")
	}

	// Check message contains the label
	if got := ctx.GetLabel("NEW_LABEL"); got != "new_value" {
		t.Errorf("expected label to be in message, got %q", got)
	}

	// Add another value for the same label
	ctx.AddLabel("NEW_LABEL", "second_value", "=")
	values := ctx.GetAllLabels("NEW_LABEL")
	if len(values) != 2 {
		t.Errorf("expected 2 values, got %d", len(values))
	}
}

func TestRemoveLabel(t *testing.T) {
	ctx := transform.NewContext("/tmp")
	ctx.Labels["LABEL"] = []string{"value1", "value2"}
	ctx.Message = "Message\nLABEL=value1\nLABEL=value2\nEnd"

	ctx.RemoveLabel("LABEL")

	// Check labels map is cleared
	if _, ok := ctx.Labels["LABEL"]; ok {
		t.Error("expected label to be removed from Labels map")
	}

	// Check message doesn't contain the label
	if got := ctx.GetLabel("LABEL"); got != "" {
		t.Errorf("expected label to be removed from message, got %q", got)
	}
}

func TestRemoveLabelWithValue(t *testing.T) {
	ctx := transform.NewContext("/tmp")
	ctx.Labels["LABEL"] = []string{"value1", "value2", "value3"}
	ctx.Message = "Message\nLABEL=value1\nLABEL=value2\nLABEL=value3\n"

	ctx.RemoveLabelWithValue("LABEL", "value2")

	// Check labels map only has value1 and value3
	values := ctx.Labels["LABEL"]
	if len(values) != 2 {
		t.Errorf("expected 2 values, got %d", len(values))
	}

	// Check value2 is not in the values
	for _, v := range values {
		if v == "value2" {
			t.Error("expected value2 to be removed")
		}
	}
}

func TestLabelPattern(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		label    string
		expected string
	}{
		{
			name:     "equals separator",
			message:  "LABEL=value",
			label:    "LABEL",
			expected: "value",
		},
		{
			name:     "colon separator",
			message:  "LABEL: value",
			label:    "LABEL",
			expected: "value",
		},
		{
			name:     "label with hyphen",
			message:  "MY-LABEL=value",
			label:    "MY-LABEL",
			expected: "value",
		},
		{
			name:     "label with underscore",
			message:  "MY_LABEL=value",
			label:    "MY_LABEL",
			expected: "value",
		},
		{
			name:     "label with digits",
			message:  "LABEL123=value",
			label:    "LABEL123",
			expected: "value",
		},
		{
			name:     "multiline message",
			message:  "First line\nLABEL=value\nLast line",
			label:    "LABEL",
			expected: "value",
		},
		{
			name:     "value with spaces",
			message:  "LABEL=value with spaces",
			label:    "LABEL",
			expected: "value with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := transform.NewContext("/tmp")
			ctx.Message = tt.message

			got := ctx.GetLabel(tt.label)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

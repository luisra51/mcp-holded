package internal

import (
	"reflect"
	"testing"
)

func TestFilterFields(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		fields   []string
		expected any
	}{
		{
			name:     "empty fields returns payload unchanged",
			input:    map[string]any{"a": 1, "b": 2},
			fields:   []string{},
			expected: map[string]any{"a": 1, "b": 2},
		},
		{
			name:   "array of maps filters each element",
			input:  []any{map[string]any{"id": "1", "name": "Alice", "email": "alice@example.com"}, map[string]any{"id": "2", "name": "Bob", "email": "bob@example.com"}},
			fields: []string{"id", "name"},
			expected: []any{
				map[string]any{"id": "1", "name": "Alice"},
				map[string]any{"id": "2", "name": "Bob"},
			},
		},
		{
			name:     "top-level map filters to requested keys",
			input:    map[string]any{"id": "123", "name": "Test", "extra": "data"},
			fields:   []string{"id", "name"},
			expected: map[string]any{"id": "123", "name": "Test"},
		},
		{
			name:     "non-existent fields simply don't appear",
			input:    map[string]any{"id": "123", "name": "Test"},
			fields:   []string{"id", "name", "nonexistent"},
			expected: map[string]any{"id": "123", "name": "Test"},
		},
		{
			name:     "scalar string returned as-is",
			input:    "hello",
			fields:   []string{"anything"},
			expected: "hello",
		},
		{
			name:     "scalar float64 returned as-is",
			input:    42.5,
			fields:   []string{"anything"},
			expected: 42.5,
		},
		{
			name:   "nested maps inside map are not filtered",
			input:  map[string]any{"id": "1", "metadata": map[string]any{"nested": "value", "other": "data"}},
			fields: []string{"id", "metadata"},
			expected: map[string]any{
				"id": "1",
				"metadata": map[string]any{"nested": "value", "other": "data"},
			},
		},
		{
			name:   "nested arrays inside array of maps",
			input:  []any{map[string]any{"id": "1", "items": []string{"a", "b"}}, map[string]any{"id": "2", "items": []string{"c", "d"}}},
			fields: []string{"id", "items"},
			expected: []any{
				map[string]any{"id": "1", "items": []string{"a", "b"}},
				map[string]any{"id": "2", "items": []string{"c", "d"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterFields(tt.input, tt.fields)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FilterFields() = %v, want %v", result, tt.expected)
			}
		})
	}
}

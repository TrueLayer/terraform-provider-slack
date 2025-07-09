package slack

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestSchemaSetToSlice(t *testing.T) {
	tests := []struct {
		name     string
		set      *schema.Set
		expected []string
	}{
		{
			name:     "empty set",
			set:      schema.NewSet(schema.HashString, []interface{}{}),
			expected: []string{},
		},
		{
			name:     "single item",
			set:      schema.NewSet(schema.HashString, []interface{}{"item1"}),
			expected: []string{"item1"},
		},
		{
			name:     "multiple items",
			set:      schema.NewSet(schema.HashString, []interface{}{"item1", "item2", "item3"}),
			expected: []string{"item1", "item2", "item3"},
		},
		{
			name:     "nil set",
			set:      nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := schemaSetToSlice(tt.set)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		remove   string
		expected []string
	}{
		{
			name:     "remove from middle",
			slice:    []string{"a", "b", "c", "d"},
			remove:   "b",
			expected: []string{"a", "c", "d"},
		},
		{
			name:     "remove from beginning",
			slice:    []string{"a", "b", "c"},
			remove:   "a",
			expected: []string{"b", "c"},
		},
		{
			name:     "remove from end",
			slice:    []string{"a", "b", "c"},
			remove:   "c",
			expected: []string{"a", "b"},
		},
		{
			name:     "remove non-existent",
			slice:    []string{"a", "b", "c"},
			remove:   "d",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "remove from empty slice",
			slice:    []string{},
			remove:   "a",
			expected: []string{},
		},
		{
			name:     "remove duplicate",
			slice:    []string{"a", "b", "a", "c"},
			remove:   "a",
			expected: []string{"b", "c"},
		},
		{
			name:     "remove all items",
			slice:    []string{"a"},
			remove:   "a",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := remove(tt.slice, tt.remove)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		element  string
		expected bool
	}{
		{
			name:     "contains element",
			slice:    []string{"a", "b", "c"},
			element:  "b",
			expected: true,
		},
		{
			name:     "does not contain element",
			slice:    []string{"a", "b", "c"},
			element:  "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			element:  "a",
			expected: false,
		},
		{
			name:     "single element match",
			slice:    []string{"a"},
			element:  "a",
			expected: true,
		},
		{
			name:     "case sensitive",
			slice:    []string{"a", "B", "c"},
			element:  "b",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.element)
			assert.Equal(t, tt.expected, result)
		})
	}
}

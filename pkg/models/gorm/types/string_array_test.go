// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with RapidaAI Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package gorm_types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    StringArray
		expected interface{} // can be nil or []byte
	}{
		{
			name:     "empty array",
			input:    StringArray{},
			expected: []byte(`[]`),
		},
		{
			name:     "non-empty array",
			input:    StringArray{"item1", "item2", "item3"},
			expected: []byte(`["item1","item2","item3"]`),
		},
		{
			name:     "nil array",
			input:    nil,
			expected: []byte(`null`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, val)
		})
	}
}

func TestStringArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected StringArray
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			hasError: false,
		},
		{
			name:     "valid JSON bytes",
			input:    []byte(`["item1","item2"]`),
			expected: StringArray{"item1", "item2"},
			hasError: false,
		},
		{
			name:     "valid JSON string",
			input:    `["a","b","c"]`,
			expected: StringArray{"a", "b", "c"},
			hasError: false,
		},
		{
			name:     "empty array JSON",
			input:    []byte(`[]`),
			expected: StringArray{},
			hasError: false,
		},
		{
			name:     "invalid JSON",
			input:    []byte(`invalid`),
			expected: nil,
			hasError: true,
		},
		{
			name:     "unsupported type",
			input:    123,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sa StringArray
			err := sa.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, sa)
			}
		})
	}
}

func TestStringArray_String(t *testing.T) {
	tests := []struct {
		name     string
		input    StringArray
		expected string
	}{
		{
			name:     "empty array",
			input:    StringArray{},
			expected: "{}",
		},
		{
			name:     "single item",
			input:    StringArray{"item"},
			expected: "{item}",
		},
		{
			name:     "multiple items",
			input:    StringArray{"a", "b", "c"},
			expected: "{a,b,c}",
		},
		{
			name:     "nil array",
			input:    nil,
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringArray_JSONMarshaling(t *testing.T) {
	sa := StringArray{"item1", "item2", "item3"}

	data, err := json.Marshal(sa)
	assert.NoError(t, err)

	var unmarshaled StringArray
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, sa, unmarshaled)
}

func TestStringArray_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var sa StringArray
		err := sa.Scan("")
		assert.Error(t, err) // empty string is not valid JSON
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var sa StringArray
		err := sa.Scan([]byte{})
		assert.Error(t, err) // empty bytes is not valid JSON
	})

	t.Run("value with special characters", func(t *testing.T) {
		sa := StringArray{"item with spaces", "item-with-dashes"}
		val, err := sa.Value()
		assert.NoError(t, err)
		expected := []byte(`["item with spaces","item-with-dashes"]`)
		assert.Equal(t, expected, val)
	})
}

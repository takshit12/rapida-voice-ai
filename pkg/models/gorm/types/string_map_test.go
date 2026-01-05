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

func TestStringMap_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    StringMap
		expected interface{} // can be nil or []byte
	}{
		{
			name:     "empty map",
			input:    StringMap{},
			expected: []byte("{}"),
		},
		{
			name: "non-empty map",
			input: StringMap{
				"key1": "value1",
				"key2": "value2",
			},
			expected: []byte(`{"key1":"value1","key2":"value2"}`),
		},
		{
			name:     "nil map",
			input:    nil,
			expected: []byte("null"),
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

func TestStringMap_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected StringMap
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: StringMap{},
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key1":"value1","key2":"value2"}`),
			expected: StringMap{
				"key1": "value1",
				"key2": "value2",
			},
			hasError: false,
		},
		{
			name:  "valid JSON string",
			input: `{"key3":"value3"}`,
			expected: StringMap{
				"key3": "value3",
			},
			hasError: false,
		},
		{
			name:     "empty JSON",
			input:    []byte(`{}`),
			expected: StringMap{},
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
			var sm StringMap
			err := sm.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, sm)
			}
		})
	}
}

func TestStringMap_String(t *testing.T) {
	tests := []struct {
		name     string
		input    StringMap
		expected string
	}{
		{
			name:     "empty map",
			input:    StringMap{},
			expected: "{}",
		},
		{
			name: "single key-value",
			input: StringMap{
				"key": "value",
			},
			expected: "{key=value}",
		},
		{
			name: "multiple key-values",
			input: StringMap{
				"a": "1",
				"b": "2",
			},
			expected: "{a:1,b:2}", // Note: order may vary, but for test we'll check contains
		},
		{
			name:     "nil map",
			input:    nil,
			expected: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			if tt.name == "multiple key-values" {
				assert.Contains(t, result, "a=1")
				assert.Contains(t, result, "b=2")
				assert.Contains(t, result, "{")
				assert.Contains(t, result, "}")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestStringMap_JSONMarshaling(t *testing.T) {
	sm := StringMap{
		"key1": "value1",
		"key2": "value2",
	}

	data, err := json.Marshal(sm)
	assert.NoError(t, err)

	var unmarshaled StringMap
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, sm, unmarshaled)
}

func TestStringMap_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var sm StringMap
		err := sm.Scan("")
		assert.NoError(t, err)
		assert.Equal(t, StringMap{}, sm)
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var sm StringMap
		err := sm.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, StringMap{}, sm)
	})

	t.Run("value with special characters", func(t *testing.T) {
		sm := StringMap{
			"key with spaces": "value with spaces",
			"key-with-dashes": "value-with-dashes",
		}
		val, err := sm.Value()
		assert.NoError(t, err)
		expected := []byte(`{"key with spaces":"value with spaces","key-with-dashes":"value-with-dashes"}`)
		assert.Equal(t, expected, val)
	})
}

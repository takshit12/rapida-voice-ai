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

func TestInterfaceMap_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    InterfaceMap
		expected string
	}{
		{
			name:     "empty map",
			input:    InterfaceMap{},
			expected: "{}",
		},
		{
			name: "non-empty map",
			input: InterfaceMap{
				"key1": "value1",
				"key2": float64(42),
				"key3": true,
			},
			expected: `{"key1":"value1","key2":42,"key3":true}`,
		},
		{
			name:     "nil map",
			input:    nil,
			expected: "null",
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

func TestInterfaceMap_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected InterfaceMap
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: InterfaceMap{},
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key1":"value1","key2":42,"key3":true}`),
			expected: InterfaceMap{
				"key1": "value1",
				"key2": float64(42),
				"key3": true,
			},
			hasError: false,
		},
		{
			name:  "valid JSON string",
			input: `{"key4":"value4"}`,
			expected: InterfaceMap{
				"key4": "value4",
			},
			hasError: false,
		},
		{
			name:     "empty JSON",
			input:    []byte(`{}`),
			expected: InterfaceMap{},
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
			var im InterfaceMap
			err := im.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, im)
			}
		})
	}
}

func TestInterfaceMap_String(t *testing.T) {
	tests := []struct {
		name     string
		input    InterfaceMap
		expected string
	}{
		{
			name:     "empty map",
			input:    InterfaceMap{},
			expected: "{}",
		},
		{
			name: "single key-value",
			input: InterfaceMap{
				"key": "value",
			},
			expected: "{key:value}",
		},
		{
			name: "multiple key-values with different types",
			input: InterfaceMap{
				"string": "text",
				"number": float64(42),
				"bool":   true,
			},
			expected: "{string:text,number:42,bool:true}", // order may vary
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
			if tt.name == "multiple key-values with different types" {
				assert.Contains(t, result, "string:text")
				assert.Contains(t, result, "number:42")
				assert.Contains(t, result, "bool:true")
				assert.Contains(t, result, "{")
				assert.Contains(t, result, "}")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestInterfaceMap_JSONMarshaling(t *testing.T) {
	im := InterfaceMap{
		"key1": "value1",
		"key2": float64(42),
		"key3": true,
		"key4": []interface{}{"a", "b"},
	}

	data, err := json.Marshal(im)
	assert.NoError(t, err)

	var unmarshaled InterfaceMap
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, im, unmarshaled)
}

func TestInterfaceMap_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var im InterfaceMap
		err := im.Scan("")
		assert.NoError(t, err)
		assert.Equal(t, InterfaceMap{}, im)
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var im InterfaceMap
		err := im.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, InterfaceMap{}, im)
	})

	t.Run("value with complex data", func(t *testing.T) {
		im := InterfaceMap{
			"nested": map[string]interface{}{
				"inner": "value",
			},
			"array": []interface{}{"a", 1, true},
		}
		val, err := im.Value()
		assert.NoError(t, err)
		expected := `{"array":["a",1,true],"nested":{"inner":"value"}}`
		assert.Equal(t, expected, val)
	})

	t.Run("scan with null values", func(t *testing.T) {
		var im InterfaceMap
		err := im.Scan([]byte(`{"key": null}`))
		assert.NoError(t, err)
		assert.Equal(t, InterfaceMap{"key": nil}, im)
	})
}

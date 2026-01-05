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

func TestIntArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    IntArray
		expected interface{} // can be nil or string
	}{
		{
			name:     "empty array",
			input:    IntArray{},
			expected: nil,
		},
		{
			name:     "non-empty array",
			input:    IntArray{1, 2, 3},
			expected: []byte(`[1,2,3]`),
		},
		{
			name:     "nil array",
			input:    nil,
			expected: nil,
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

func TestIntArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected IntArray
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: IntArray{},
			hasError: false,
		},
		{
			name:     "valid JSON bytes",
			input:    []byte(`[1,2,3]`),
			expected: IntArray{1, 2, 3},
			hasError: false,
		},
		{
			name:     "valid JSON string",
			input:    `[10,20,30]`,
			expected: IntArray{10, 20, 30},
			hasError: false,
		},
		{
			name:     "empty array JSON",
			input:    []byte(`[]`),
			expected: IntArray{},
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
			var ia IntArray
			err := ia.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, ia)
			}
		})
	}
}

func TestIntArray_String(t *testing.T) {
	tests := []struct {
		name     string
		input    IntArray
		expected string
	}{
		{
			name:     "empty array",
			input:    IntArray{},
			expected: "{}",
		},
		{
			name:     "single item",
			input:    IntArray{42},
			expected: "{42}",
		},
		{
			name:     "multiple items",
			input:    IntArray{1, 2, 3},
			expected: "{1,2,3}",
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

func TestIntArray_JSONMarshaling(t *testing.T) {
	ia := IntArray{1, 2, 3, 4, 5}

	data, err := json.Marshal(ia)
	assert.NoError(t, err)

	var unmarshaled IntArray
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, ia, unmarshaled)
}

func TestIntArray_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var ia IntArray
		err := ia.Scan("")
		assert.NoError(t, err)
		assert.Equal(t, IntArray{}, ia)
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var ia IntArray
		err := ia.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, IntArray{}, ia)
	})

	t.Run("value with large numbers", func(t *testing.T) {
		ia := IntArray{9223372036854775807, 0, 1} // max uint64 and some values
		val, err := ia.Value()
		assert.NoError(t, err)
		expected := `[9223372036854775807,0,1]`
		assert.Equal(t, []byte(expected), val)
	})

	t.Run("string representation", func(t *testing.T) {
		ia := IntArray{100, 200, 300}
		str := ia.String()
		assert.Equal(t, "{100,200,300}", str)
	})
}

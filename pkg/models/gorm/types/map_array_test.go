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

func TestMapArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    MapArray
		expected interface{} // can be nil or string
	}{
		{
			name:     "empty array",
			input:    MapArray{},
			expected: nil,
		},
		{
			name: "non-empty array",
			input: MapArray{
				{"key1": "value1", "key2": "value2"},
				{"key3": "value3"},
			},
			expected: `[{"key1":"value1","key2":"value2"},{"key3":"value3"}]`,
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

func TestMapArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected MapArray
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: MapArray{},
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`[{"key1":"value1"},{"key2":"value2"}]`),
			expected: MapArray{
				{"key1": "value1"},
				{"key2": "value2"},
			},
			hasError: false,
		},
		{
			name:  "valid JSON string",
			input: `[{"a":"1"},{"b":"2"}]`,
			expected: MapArray{
				{"a": "1"},
				{"b": "2"},
			},
			hasError: false,
		},
		{
			name:     "empty array JSON",
			input:    []byte(`[]`),
			expected: MapArray{},
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
			var ma MapArray
			err := ma.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, ma)
			}
		})
	}
}

func TestMapArray_String(t *testing.T) {
	tests := []struct {
		name  string
		input MapArray
	}{
		{
			name:  "empty array",
			input: MapArray{},
		},
		{
			name: "non-empty array",
			input: MapArray{
				{"key": "value"},
			},
		},
		{
			name:  "nil array",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			// Since String() marshals to JSON, it should be valid JSON
			var unmarshaled MapArray
			err := json.Unmarshal([]byte(result), &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}

func TestMapInterfaceArray_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    MapInterfaceArray
		expected interface{}
	}{
		{
			name:     "empty array",
			input:    MapInterfaceArray{},
			expected: nil,
		},
		{
			name: "non-empty array",
			input: MapInterfaceArray{
				{"key1": "value1", "num": float64(42)},
				{"key2": true},
			},
			expected: `[{"key1":"value1","num":42},{"key2":true}]`,
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

func TestMapInterfaceArray_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected MapInterfaceArray
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: MapInterfaceArray{},
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`[{"key1":"value1","num":42},{"key2":true}]`),
			expected: MapInterfaceArray{
				{"key1": "value1", "num": float64(42)},
				{"key2": true},
			},
			hasError: false,
		},
		{
			name:  "valid JSON string",
			input: `[{"a":"1"},{"b":2}]`,
			expected: MapInterfaceArray{
				{"a": "1"},
				{"b": float64(2)},
			},
			hasError: false,
		},
		{
			name:     "empty array JSON",
			input:    []byte(`[]`),
			expected: MapInterfaceArray{},
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
			var mia MapInterfaceArray
			err := mia.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, mia)
			}
		})
	}
}

func TestMapInterfaceArray_String(t *testing.T) {
	tests := []struct {
		name  string
		input MapInterfaceArray
	}{
		{
			name:  "empty array",
			input: MapInterfaceArray{},
		},
		{
			name: "non-empty array",
			input: MapInterfaceArray{
				{"key": "value", "num": float64(42)},
			},
		},
		{
			name:  "nil array",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			// Since String() marshals to JSON, it should be valid JSON
			var unmarshaled MapInterfaceArray
			err := json.Unmarshal([]byte(result), &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, unmarshaled)
		})
	}
}

func TestMapArray_JSONMarshaling(t *testing.T) {
	ma := MapArray{
		{"key1": "value1"},
		{"key2": "value2"},
	}

	data, err := json.Marshal(ma)
	assert.NoError(t, err)

	var unmarshaled MapArray
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, ma, unmarshaled)
}

func TestMapInterfaceArray_JSONMarshaling(t *testing.T) {
	mia := MapInterfaceArray{
		{"key1": "value1", "num": float64(42)},
		{"key2": true},
	}

	data, err := json.Marshal(mia)
	assert.NoError(t, err)

	var unmarshaled MapInterfaceArray
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, mia, unmarshaled)
}

func TestMapArray_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var ma MapArray
		err := ma.Scan("")
		assert.NoError(t, err)
		assert.Equal(t, MapArray{}, ma)
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var ma MapArray
		err := ma.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, MapArray{}, ma)
	})

	t.Run("value with nested structures", func(t *testing.T) {
		ma := MapArray{
			{
				"simple": "value",
				"number": "42",
			},
		}
		val, err := ma.Value()
		assert.NoError(t, err)
		expected := `[{"number":"42","simple":"value"}]`
		assert.Equal(t, expected, val)
	})
}

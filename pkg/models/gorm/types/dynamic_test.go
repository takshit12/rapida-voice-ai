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

func TestNewDynamic(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "string",
			input: "test string",
		},
		{
			name:  "int",
			input: 42,
		},
		{
			name:  "float64",
			input: 3.14,
		},
		{
			name:  "bool",
			input: true,
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value"},
		},
		{
			name:  "slice",
			input: []string{"a", "b", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDynamic(tt.input)
			assert.Equal(t, tt.input, d.Data)
		})
	}
}

func TestDynamic_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string",
			input:    "test",
			expected: "test",
		},
		{
			name:     "int",
			input:    42,
			expected: 42,
		},
		{
			name:     "float64",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "bool",
			input:    true,
			expected: true,
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key": "value",
				"num": 42,
			},
			expected: []byte(`{"key":"value","num":42}`),
		},
		{
			name:     "slice",
			input:    []interface{}{"a", 1, true},
			expected: []byte(`["a",1,true]`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dynamic{Data: tt.input}
			val, err := d.Value()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, val)
		})
	}
}

func TestDynamic_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
			hasError: false,
		},
		{
			name:     "string input",
			input:    "test string",
			expected: "test string",
			hasError: false,
		},
		{
			name:     "int64 input",
			input:    int64(42),
			expected: 42,
			hasError: false,
		},
		{
			name:     "float64 input",
			input:    3.14,
			expected: 3.14,
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key":"value","num":42}`),
			expected: map[string]interface{}{
				"key": "value",
				"num": float64(42),
			},
			hasError: false,
		},
		{
			name:     "invalid JSON bytes",
			input:    []byte(`invalid json`),
			expected: "invalid json", // fallback to string
			hasError: false,
		},
		{
			name:     "unsupported type",
			input:    make(chan int),
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Dynamic
			err := d.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, d.Data)
			}
		})
	}
}

func TestDynamic_Get(t *testing.T) {
	d := Dynamic{Data: "test value"}
	assert.Equal(t, "test value", d.Get())
}

func TestDynamic_GetString(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected string
		ok       bool
	}{
		{
			name:     "string data",
			data:     "hello",
			expected: "hello",
			ok:       true,
		},
		{
			name:     "non-string data",
			data:     42,
			expected: "",
			ok:       false,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: "",
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dynamic{Data: tt.data}
			result, ok := d.GetString()
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestDynamic_GetInt(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected int
		ok       bool
	}{
		{
			name:     "int data",
			data:     42,
			expected: 42,
			ok:       true,
		},
		{
			name:     "non-int data",
			data:     "string",
			expected: 0,
			ok:       false,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dynamic{Data: tt.data}
			result, ok := d.GetInt()
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestDynamic_GetMap(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected map[string]interface{}
		ok       bool
	}{
		{
			name: "map data",
			data: map[string]interface{}{
				"key": "value",
			},
			expected: map[string]interface{}{
				"key": "value",
			},
			ok: true,
		},
		{
			name:     "non-map data",
			data:     "string",
			expected: nil,
			ok:       false,
		},
		{
			name:     "nil data",
			data:     nil,
			expected: nil,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Dynamic{Data: tt.data}
			result, ok := d.GetMap()
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.ok, ok)
		})
	}
}

func TestDynamic_JSONMarshaling(t *testing.T) {
	original := Dynamic{Data: map[string]interface{}{
		"key":  "value",
		"num":  float64(42),
		"bool": true,
	}}

	data, err := json.Marshal(original)
	assert.NoError(t, err)

	var unmarshaled Dynamic
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, original.Data, unmarshaled.Data)
}

func TestDynamic_EdgeCases(t *testing.T) {
	t.Run("scan with empty bytes", func(t *testing.T) {
		var d Dynamic
		err := d.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{}, d.Data)
	})

	t.Run("value with nil data", func(t *testing.T) {
		d := Dynamic{Data: nil}
		val, err := d.Value()
		assert.NoError(t, err)
		assert.Nil(t, val)
	})

	t.Run("scan with complex JSON", func(t *testing.T) {
		var d Dynamic
		input := []byte(`{"nested":{"key":"value"},"array":[1,2,3]}`)
		err := d.Scan(input)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"array": []interface{}{float64(1), float64(2), float64(3)},
		}
		assert.Equal(t, expected, d.Data)
	})
}

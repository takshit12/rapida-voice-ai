// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package gorm_models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type badMarshaler struct{}

func (b badMarshaler) MarshalJSON() ([]byte, error) {
	return nil, assert.AnError // some error
}

func TestArgument_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  Argument
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "string value",
			initial:  Argument{Name: "test"},
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "[]byte value",
			initial:  Argument{Name: "test"},
			input:    []byte("world"),
			expected: "world",
			wantErr:  false,
		},
		{
			name:     "int value",
			initial:  Argument{Name: "test"},
			input:    42,
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "int8 value",
			initial:  Argument{Name: "test"},
			input:    int8(8),
			expected: "8",
			wantErr:  false,
		},
		{
			name:     "int16 value",
			initial:  Argument{Name: "test"},
			input:    int16(16),
			expected: "16",
			wantErr:  false,
		},
		{
			name:     "int32 value",
			initial:  Argument{Name: "test"},
			input:    int32(32),
			expected: "32",
			wantErr:  false,
		},
		{
			name:     "int64 value",
			initial:  Argument{Name: "test"},
			input:    int64(64),
			expected: "64",
			wantErr:  false,
		},
		{
			name:     "uint value",
			initial:  Argument{Name: "test"},
			input:    uint(1),
			expected: "1",
			wantErr:  false,
		},
		{
			name:     "uint8 value",
			initial:  Argument{Name: "test"},
			input:    uint8(8),
			expected: "8",
			wantErr:  false,
		},
		{
			name:     "uint16 value",
			initial:  Argument{Name: "test"},
			input:    uint16(16),
			expected: "16",
			wantErr:  false,
		},
		{
			name:     "uint32 value",
			initial:  Argument{Name: "test"},
			input:    uint32(32),
			expected: "32",
			wantErr:  false,
		},
		{
			name:     "uint64 value",
			initial:  Argument{Name: "test"},
			input:    uint64(64),
			expected: "64",
			wantErr:  false,
		},
		{
			name:     "float32 value",
			initial:  Argument{Name: "test"},
			input:    float32(3.14),
			expected: "3.140000",
			wantErr:  false,
		},
		{
			name:     "float64 value",
			initial:  Argument{Name: "test"},
			input:    2.71,
			expected: "2.710000",
			wantErr:  false,
		},
		{
			name:     "bool true",
			initial:  Argument{Name: "test"},
			input:    true,
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "bool false",
			initial:  Argument{Name: "test"},
			input:    false,
			expected: "false",
			wantErr:  false,
		},
		{
			name:     "nil value",
			initial:  Argument{Name: "test"},
			input:    nil,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "map value",
			initial:  Argument{Name: "test"},
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "struct value",
			initial:  Argument{Name: "test"},
			input:    struct{ Field string }{"test"},
			expected: `{"Field":"test"}`,
			wantErr:  false,
		},
		{
			name:     "slice value",
			initial:  Argument{Name: "test"},
			input:    []string{"a", "b"},
			expected: `["a","b"]`,
			wantErr:  false,
		},
		{
			name:     "marshal error",
			initial:  Argument{Name: "test"},
			input:    badMarshaler{},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arg := tt.initial
			err := arg.SetValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, arg.Value)
				assert.Equal(t, tt.initial.Name, arg.Name) // Name should remain unchanged
			}
		})
	}
}

func TestArgument_JSONMarshaling(t *testing.T) {
	// Test that Argument can be JSON marshaled
	arg := Argument{
		Name:  "test_arg",
		Value: "test_value",
	}

	data, err := json.Marshal(arg)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"name":"test_arg"`)
	assert.Contains(t, string(data), `"value":"test_value"`)
}

func TestArgument_JSONUnmarshaling(t *testing.T) {
	// Test that Argument can be JSON unmarshaled
	jsonStr := `{"name":"test_arg","value":"test_value"}`
	var arg Argument
	err := json.Unmarshal([]byte(jsonStr), &arg)
	assert.NoError(t, err)
	assert.Equal(t, "test_arg", arg.Name)
	assert.Equal(t, "test_value", arg.Value)
}

func TestArgument_EdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		arg := Argument{Name: "test"}
		err := arg.SetValue("")
		assert.NoError(t, err)
		assert.Equal(t, "", arg.Value)
	})

	t.Run("zero values", func(t *testing.T) {
		arg := Argument{Name: "test"}
		err := arg.SetValue(0)
		assert.NoError(t, err)
		assert.Equal(t, "0", arg.Value)
	})

	t.Run("large numbers", func(t *testing.T) {
		arg := Argument{Name: "test"}
		err := arg.SetValue(int64(9223372036854775807))
		assert.NoError(t, err)
		assert.Equal(t, "9223372036854775807", arg.Value)
	})

	t.Run("negative numbers", func(t *testing.T) {
		arg := Argument{Name: "test"}
		err := arg.SetValue(-42)
		assert.NoError(t, err)
		assert.Equal(t, "-42", arg.Value)
	})

	t.Run("complex map", func(t *testing.T) {
		arg := Argument{Name: "test"}
		input := map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"array": []int{1, 2, 3},
		}
		err := arg.SetValue(input)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal([]byte(arg.Value), &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["nested"].(map[string]interface{})["key"])
	})
}

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

type badMarshalerMeta struct{}

func (b badMarshalerMeta) MarshalJSON() ([]byte, error) {
	return nil, assert.AnError
}

func TestNewMetadata(t *testing.T) {
	key := "test_key"
	value := "test_value"

	metadata := NewMetadata(key, value)

	assert.NotNil(t, metadata)
	assert.Equal(t, key, metadata.Key)
	assert.Equal(t, value, metadata.Value)
}

func TestMetadata_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  Metadata
		input    interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "string value",
			initial:  Metadata{Key: "test"},
			input:    "hello",
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "[]byte value",
			initial:  Metadata{Key: "test"},
			input:    []byte("world"),
			expected: "world",
			wantErr:  false,
		},
		{
			name:     "int value",
			initial:  Metadata{Key: "test"},
			input:    42,
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "int8 value",
			initial:  Metadata{Key: "test"},
			input:    int8(8),
			expected: "8",
			wantErr:  false,
		},
		{
			name:     "int16 value",
			initial:  Metadata{Key: "test"},
			input:    int16(16),
			expected: "16",
			wantErr:  false,
		},
		{
			name:     "int32 value",
			initial:  Metadata{Key: "test"},
			input:    int32(32),
			expected: "32",
			wantErr:  false,
		},
		{
			name:     "int64 value",
			initial:  Metadata{Key: "test"},
			input:    int64(64),
			expected: "64",
			wantErr:  false,
		},
		{
			name:     "uint value",
			initial:  Metadata{Key: "test"},
			input:    uint(1),
			expected: "1",
			wantErr:  false,
		},
		{
			name:     "uint8 value",
			initial:  Metadata{Key: "test"},
			input:    uint8(8),
			expected: "8",
			wantErr:  false,
		},
		{
			name:     "uint16 value",
			initial:  Metadata{Key: "test"},
			input:    uint16(16),
			expected: "16",
			wantErr:  false,
		},
		{
			name:     "uint32 value",
			initial:  Metadata{Key: "test"},
			input:    uint32(32),
			expected: "32",
			wantErr:  false,
		},
		{
			name:     "uint64 value",
			initial:  Metadata{Key: "test"},
			input:    uint64(64),
			expected: "64",
			wantErr:  false,
		},
		{
			name:     "float32 value",
			initial:  Metadata{Key: "test"},
			input:    float32(3.14),
			expected: "3.140000",
			wantErr:  false,
		},
		{
			name:     "float64 value",
			initial:  Metadata{Key: "test"},
			input:    2.71,
			expected: "2.710000",
			wantErr:  false,
		},
		{
			name:     "bool true",
			initial:  Metadata{Key: "test"},
			input:    true,
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "bool false",
			initial:  Metadata{Key: "test"},
			input:    false,
			expected: "false",
			wantErr:  false,
		},
		{
			name:     "nil value",
			initial:  Metadata{Key: "test"},
			input:    nil,
			expected: "",
			wantErr:  false,
		},
		{
			name:     "map value",
			initial:  Metadata{Key: "test"},
			input:    map[string]interface{}{"key": "value"},
			expected: `{"key":"value"}`,
			wantErr:  false,
		},
		{
			name:     "struct value",
			initial:  Metadata{Key: "test"},
			input:    struct{ Field string }{"test"},
			expected: `{"Field":"test"}`,
			wantErr:  false,
		},
		{
			name:     "slice value",
			initial:  Metadata{Key: "test"},
			input:    []string{"a", "b"},
			expected: `["a","b"]`,
			wantErr:  false,
		},
		{
			name:     "marshal error",
			initial:  Metadata{Key: "test"},
			input:    badMarshalerMeta{},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := tt.initial
			err := md.SetValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, md.Value)
				assert.Equal(t, tt.initial.Key, md.Key) // Key should remain unchanged
			}
		})
	}
}

func TestMetadata_JSONMarshaling(t *testing.T) {
	md := Metadata{
		Key:   "test_key",
		Value: "test_value",
	}

	data, err := json.Marshal(md)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"key":"test_key"`)
	assert.Contains(t, string(data), `"value":"test_value"`)
}

func TestMetadata_JSONUnmarshaling(t *testing.T) {
	jsonStr := `{"key":"test_key","value":"test_value"}`
	var md Metadata
	err := json.Unmarshal([]byte(jsonStr), &md)
	assert.NoError(t, err)
	assert.Equal(t, "test_key", md.Key)
	assert.Equal(t, "test_value", md.Value)
}

func TestMetadata_EdgeCases(t *testing.T) {
	t.Run("empty strings", func(t *testing.T) {
		md := NewMetadata("", "")
		assert.Equal(t, "", md.Key)
		assert.Equal(t, "", md.Value)
	})

	t.Run("zero values", func(t *testing.T) {
		md := Metadata{Key: "test"}
		err := md.SetValue(0)
		assert.NoError(t, err)
		assert.Equal(t, "0", md.Value)
	})

	t.Run("large numbers", func(t *testing.T) {
		md := Metadata{Key: "test"}
		err := md.SetValue(int64(9223372036854775807))
		assert.NoError(t, err)
		assert.Equal(t, "9223372036854775807", md.Value)
	})

	t.Run("negative numbers", func(t *testing.T) {
		md := Metadata{Key: "test"}
		err := md.SetValue(-42)
		assert.NoError(t, err)
		assert.Equal(t, "-42", md.Value)
	})

	t.Run("complex map", func(t *testing.T) {
		md := Metadata{Key: "test"}
		input := map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"array": []int{1, 2, 3},
		}
		err := md.SetValue(input)
		assert.NoError(t, err)
		var result map[string]interface{}
		err = json.Unmarshal([]byte(md.Value), &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result["nested"].(map[string]interface{})["key"])
	})
}

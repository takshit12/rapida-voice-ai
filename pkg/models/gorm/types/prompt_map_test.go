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

func TestPromptMap_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    PromptMap
		expected string
	}{
		{
			name:     "empty map",
			input:    PromptMap{},
			expected: "{}",
		},
		{
			name: "non-empty map",
			input: PromptMap{
				"key1": "value1",
				"key2": float64(42),
			},
			expected: `{"key1":"value1","key2":42}`,
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

func TestPromptMap_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected PromptMap
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: PromptMap{},
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key1":"value1","key2":42}`),
			expected: PromptMap{
				"key1": "value1",
				"key2": float64(42),
			},
			hasError: false,
		},
		{
			name:  "valid JSON string",
			input: `{"key3":"value3"}`,
			expected: PromptMap{
				"key3": "value3",
			},
			hasError: false,
		},
		{
			name:     "empty JSON",
			input:    []byte(`{}`),
			expected: PromptMap{},
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
			var pm PromptMap
			err := pm.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, pm)
			}
		})
	}
}

func TestPromptMap_JSONMarshaling(t *testing.T) {
	pm := PromptMap{
		"key1": "value1",
		"key2": float64(42),
		"key3": true,
	}

	data, err := json.Marshal(pm)
	assert.NoError(t, err)

	var unmarshaled PromptMap
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, pm, unmarshaled)
}

func TestPromptMap_GetTextChatCompleteTemplate(t *testing.T) {
	tests := []struct {
		name     string
		input    PromptMap
		expected *TextChatCompletePromptTemplate
	}{
		{
			name: "valid template",
			input: PromptMap{
				"prompt": []interface{}{
					map[string]interface{}{
						"role":    "system",
						"content": "You are a helpful assistant",
					},
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
				"promptVariables": []interface{}{
					map[string]interface{}{
						"name":         "var1",
						"type":         "string",
						"defaultValue": "default",
					},
				},
			},
			expected: &TextChatCompletePromptTemplate{
				Prompt: []*PromptTemplate{
					{
						Role:    "system",
						Content: "You are a helpful assistant",
					},
					{
						Role:    "user",
						Content: "Hello",
					},
				},
				Variables: []*PromptVariable{
					{
						Name:         "var1",
						Type:         "string",
						DefaultValue: "default",
					},
				},
			},
		},
		{
			name: "invalid template - no variables",
			input: PromptMap{
				"prompt": []interface{}{
					map[string]interface{}{
						"role":    "user",
						"content": "Hello",
					},
				},
			},
			expected: nil, // because len(Variables) == 0
		},
		{
			name: "invalid JSON",
			input: PromptMap{
				"invalid": make(chan int), // unmarshalable
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.GetTextChatCompleteTemplate()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Prompt, result.Prompt)
				assert.Equal(t, tt.expected.Variables, result.Variables)
			}
		})
	}
}

func TestPromptTemplate_GetRole(t *testing.T) {
	pt := &PromptTemplate{
		Role:    "assistant",
		Content: "Response",
	}
	assert.Equal(t, "assistant", pt.GetRole())
}

func TestPromptTemplate_GetContent(t *testing.T) {
	pt := &PromptTemplate{
		Role:    "user",
		Content: "Question",
	}
	assert.Equal(t, "Question", pt.GetContent())
}

func TestPromptMap_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var pm PromptMap
		err := pm.Scan("")
		assert.NoError(t, err)
		assert.Equal(t, PromptMap{}, pm)
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var pm PromptMap
		err := pm.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, PromptMap{}, pm)
	})

	t.Run("value with complex data", func(t *testing.T) {
		pm := PromptMap{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"array": []string{"a", "b"},
		}
		val, err := pm.Value()
		assert.NoError(t, err)
		expected := `{"array":["a","b"],"nested":{"key":"value"}}`
		assert.Equal(t, expected, val)
	})
}

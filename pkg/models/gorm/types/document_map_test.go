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

func TestDocumentMap_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    DocumentMap
		expected string
	}{
		{
			name:     "empty map",
			input:    DocumentMap{},
			expected: "{}",
		},
		{
			name: "non-empty map",
			input: DocumentMap{
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

func TestDocumentMap_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected DocumentMap
		hasError bool
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: DocumentMap{},
			hasError: false,
		},
		{
			name:  "valid JSON bytes",
			input: []byte(`{"key1":"value1","key2":42}`),
			expected: DocumentMap{
				"key1": "value1",
				"key2": float64(42),
			},
			hasError: false,
		},
		{
			name:  "valid JSON string",
			input: `{"key3":"value3"}`,
			expected: DocumentMap{
				"key3": "value3",
			},
			hasError: false,
		},
		{
			name:     "empty JSON",
			input:    []byte(`{}`),
			expected: DocumentMap{},
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
			var dm DocumentMap
			err := dm.Scan(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, dm)
			}
		})
	}
}

func TestDocumentMap_JSONMarshaling(t *testing.T) {
	dm := DocumentMap{
		"key1": "value1",
		"key2": float64(42),
		"key3": true,
	}

	data, err := json.Marshal(dm)
	assert.NoError(t, err)

	var unmarshaled DocumentMap
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, dm, unmarshaled)
}

func TestDocumentStruct_JSONMarshaling(t *testing.T) {
	ds := DocumentStruct{
		Source:      "manual",
		Type:        "pdf",
		DocumentUrl: "http://example.com/doc.pdf",
	}

	data, err := json.Marshal(ds)
	assert.NoError(t, err)

	var unmarshaled DocumentStruct
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, ds, unmarshaled)
}

func TestInternalDocumentStruct_JSONMarshaling(t *testing.T) {
	ids := InternalDocumentStruct{
		DocumentStruct: DocumentStruct{
			Source:      "manual",
			Type:        "pdf",
			DocumentUrl: "http://example.com/doc.pdf",
		},
		Size: 1024,
	}

	data, err := json.Marshal(ids)
	assert.NoError(t, err)

	var unmarshaled InternalDocumentStruct
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, ids, unmarshaled)
}

func TestExternalDocumentStruct_JSONMarshaling(t *testing.T) {
	eds := ExternalDocumentStruct{
		DocumentStruct: DocumentStruct{
			Source:      "github",
			Type:        "code",
			DocumentUrl: "https://github.com/user/repo",
		},
		// Additional fields would be tested here if they existed
	}

	data, err := json.Marshal(eds)
	assert.NoError(t, err)

	var unmarshaled ExternalDocumentStruct
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, eds, unmarshaled)
}

func TestDocumentMap_EdgeCases(t *testing.T) {
	t.Run("scan with empty string", func(t *testing.T) {
		var dm DocumentMap
		err := dm.Scan("")
		assert.NoError(t, err)
		assert.Equal(t, DocumentMap{}, dm)
	})

	t.Run("scan with empty bytes", func(t *testing.T) {
		var dm DocumentMap
		err := dm.Scan([]byte{})
		assert.NoError(t, err)
		assert.Equal(t, DocumentMap{}, dm)
	})

	t.Run("value with complex data", func(t *testing.T) {
		dm := DocumentMap{
			"nested": map[string]interface{}{
				"inner": "value",
			},
			"array": []interface{}{"a", 1, true},
		}
		val, err := dm.Value()
		assert.NoError(t, err)
		expected := `{"array":["a",1,true],"nested":{"inner":"value"}}`
		assert.Equal(t, expected, val)
	})

	t.Run("scan with null values", func(t *testing.T) {
		var dm DocumentMap
		err := dm.Scan([]byte(`{"key": null}`))
		assert.NoError(t, err)
		assert.Equal(t, DocumentMap{"key": nil}, dm)
	})
}

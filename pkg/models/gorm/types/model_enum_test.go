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

func TestRetrievalMethod_Value(t *testing.T) {
	tests := []struct {
		name     string
		input    RetrievalMethod
		expected string
	}{
		{
			name:     "semantic search",
			input:    RETRIEVAL_METHOD_SEMANTIC,
			expected: "semantic-search",
		},
		{
			name:     "full text search",
			input:    RETRIEVAL_METHOD_FULLTEXT,
			expected: "full-text-search",
		},
		{
			name:     "hybrid search",
			input:    RETRIEVAL_METHOD_HYBRID,
			expected: "hybrid-search",
		},
		{
			name:     "inverted index",
			input:    RETRIEVAL_METHOD_INVERTEDINDEX,
			expected: "inverted-index",
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

func TestRetrievalMethod_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    RetrievalMethod
		expected string
	}{
		{
			name:     "semantic search",
			input:    RETRIEVAL_METHOD_SEMANTIC,
			expected: `"semantic-search"`,
		},
		{
			name:     "full text search",
			input:    RETRIEVAL_METHOD_FULLTEXT,
			expected: `"full-text-search"`,
		},
		{
			name:     "hybrid search",
			input:    RETRIEVAL_METHOD_HYBRID,
			expected: `"hybrid-search"`,
		},
		{
			name:     "inverted index",
			input:    RETRIEVAL_METHOD_INVERTEDINDEX,
			expected: `"inverted-index"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.input.MarshalJSON()
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestDocumentSource_Constants(t *testing.T) {
	assert.Equal(t, DocumentSource("manual"), DOCUMENT_SOURCE_MANUAL)
}

func TestManualDocumentSource_Constants(t *testing.T) {
	assert.Equal(t, ManualDocumentSource("manual-file"), DOCUMENT_SOURCE_MANUAL_FILE)
	assert.Equal(t, ManualDocumentSource("manual-zip"), DOCUMENT_SOURCE_MANUAL_ZIP)
	assert.Equal(t, ManualDocumentSource("manual-url"), DOCUMENT_SOURCE_MANUAL_URL)
}

func TestDocumentType_Constants(t *testing.T) {
	assert.Equal(t, "pdf", DOCUMENT_TYPE_PDF)
	assert.Equal(t, "pdf", DOCUMENT_TYPE_CSV)
	assert.Equal(t, "pdf", DOCUMENT_TYPE_XLS)
	assert.Equal(t, "txt", DOCUMENT_TYPE_TXT)
	assert.Equal(t, "web-url", DOCUMENT_TYPE_WEB_URL)
	assert.Equal(t, "unknown", DOCUMENT_TYPE_UNKNOWN)
	assert.Equal(t, "source-code", DOCUMENT_TYPE_CODE)
}

func TestRetrievalMethod_JSONRoundTrip(t *testing.T) {
	methods := []RetrievalMethod{
		RETRIEVAL_METHOD_SEMANTIC,
		RETRIEVAL_METHOD_FULLTEXT,
		RETRIEVAL_METHOD_HYBRID,
		RETRIEVAL_METHOD_INVERTEDINDEX,
	}

	for _, original := range methods {
		data, err := json.Marshal(original)
		assert.NoError(t, err)

		var unmarshaled RetrievalMethod
		err = json.Unmarshal(data, &unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, original, unmarshaled)
	}
}

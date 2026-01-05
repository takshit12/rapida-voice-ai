// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package gorm_models

import (
	"encoding/json"
	"testing"

	gorm_types "github.com/rapidaai/pkg/models/gorm/types"
	"github.com/stretchr/testify/assert"
)

type badMarshalerEvent struct{}

func (b badMarshalerEvent) MarshalJSON() ([]byte, error) {
	return nil, assert.AnError
}

func TestNewEvent(t *testing.T) {
	eventType := "test_event"
	value := "test_value"

	event := NewEvent(eventType, value)

	assert.NotNil(t, event)
	assert.Equal(t, eventType, event.EventType)
	assert.Equal(t, gorm_types.InterfaceMap{"value": value}, event.Payload)
}

func TestEvent_SetValue(t *testing.T) {
	tests := []struct {
		name     string
		initial  Event
		input    interface{}
		expected gorm_types.InterfaceMap
		wantErr  bool
	}{
		{
			name:     "string value",
			initial:  Event{EventType: "test"},
			input:    "hello",
			expected: gorm_types.InterfaceMap{"value": "hello"},
			wantErr:  false,
		},
		{
			name:     "[]byte value",
			initial:  Event{EventType: "test"},
			input:    []byte("world"),
			expected: gorm_types.InterfaceMap{"value": "world"},
			wantErr:  false,
		},
		{
			name:     "int value",
			initial:  Event{EventType: "test"},
			input:    42,
			expected: gorm_types.InterfaceMap{"value": "42"},
			wantErr:  false,
		},
		{
			name:     "int8 value",
			initial:  Event{EventType: "test"},
			input:    int8(8),
			expected: gorm_types.InterfaceMap{"value": "8"},
			wantErr:  false,
		},
		{
			name:     "int16 value",
			initial:  Event{EventType: "test"},
			input:    int16(16),
			expected: gorm_types.InterfaceMap{"value": "16"},
			wantErr:  false,
		},
		{
			name:     "int32 value",
			initial:  Event{EventType: "test"},
			input:    int32(32),
			expected: gorm_types.InterfaceMap{"value": "32"},
			wantErr:  false,
		},
		{
			name:     "int64 value",
			initial:  Event{EventType: "test"},
			input:    int64(64),
			expected: gorm_types.InterfaceMap{"value": "64"},
			wantErr:  false,
		},
		{
			name:     "uint value",
			initial:  Event{EventType: "test"},
			input:    uint(1),
			expected: gorm_types.InterfaceMap{"value": "1"},
			wantErr:  false,
		},
		{
			name:     "uint8 value",
			initial:  Event{EventType: "test"},
			input:    uint8(8),
			expected: gorm_types.InterfaceMap{"value": "8"},
			wantErr:  false,
		},
		{
			name:     "uint16 value",
			initial:  Event{EventType: "test"},
			input:    uint16(16),
			expected: gorm_types.InterfaceMap{"value": "16"},
			wantErr:  false,
		},
		{
			name:     "uint32 value",
			initial:  Event{EventType: "test"},
			input:    uint32(32),
			expected: gorm_types.InterfaceMap{"value": "32"},
			wantErr:  false,
		},
		{
			name:     "uint64 value",
			initial:  Event{EventType: "test"},
			input:    uint64(64),
			expected: gorm_types.InterfaceMap{"value": "64"},
			wantErr:  false,
		},
		{
			name:     "float32 value",
			initial:  Event{EventType: "test"},
			input:    float32(3.14),
			expected: gorm_types.InterfaceMap{"value": "3.140000"},
			wantErr:  false,
		},
		{
			name:     "float64 value",
			initial:  Event{EventType: "test"},
			input:    2.71,
			expected: gorm_types.InterfaceMap{"value": "2.710000"},
			wantErr:  false,
		},
		{
			name:     "bool true",
			initial:  Event{EventType: "test"},
			input:    true,
			expected: gorm_types.InterfaceMap{"value": "true"},
			wantErr:  false,
		},
		{
			name:     "bool false",
			initial:  Event{EventType: "test"},
			input:    false,
			expected: gorm_types.InterfaceMap{"value": "false"},
			wantErr:  false,
		},
		{
			name:     "nil value",
			initial:  Event{EventType: "test"},
			input:    nil,
			expected: gorm_types.InterfaceMap{},
			wantErr:  false,
		},
		{
			name:     "map value",
			initial:  Event{EventType: "test"},
			input:    map[string]interface{}{"key": "value"},
			expected: gorm_types.InterfaceMap{"key": "value"},
			wantErr:  false,
		},
		{
			name:     "struct value",
			initial:  Event{EventType: "test"},
			input:    struct{ Field string }{"test"},
			expected: gorm_types.InterfaceMap{"Field": "test"},
			wantErr:  false,
		},
		{
			name:     "marshal error",
			initial:  Event{EventType: "test"},
			input:    badMarshalerEvent{},
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "unmarshal error",
			initial:  Event{EventType: "test"},
			input:    []int{1, 2, 3}, // marshals to array, can't unmarshal to map
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := tt.initial
			err := event.SetValue(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, event.Payload)
				assert.Equal(t, tt.initial.EventType, event.EventType) // EventType should remain unchanged
			}
		})
	}
}

func TestEvent_JSONMarshaling(t *testing.T) {
	event := Event{
		EventType: "user_login",
		Payload:   gorm_types.InterfaceMap{"user_id": 123, "ip": "192.168.1.1"},
	}

	data, err := json.Marshal(event)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"eventType":"user_login"`)
	assert.Contains(t, string(data), `"payload":{"ip":"192.168.1.1","user_id":123}`)
}

func TestEvent_JSONUnmarshaling(t *testing.T) {
	jsonStr := `{"eventType":"user_logout","payload":{"user_id":456,"session_id":"abc"}}`
	var event Event
	err := json.Unmarshal([]byte(jsonStr), &event)
	assert.NoError(t, err)
	assert.Equal(t, "user_logout", event.EventType)
	assert.Equal(t, gorm_types.InterfaceMap{"user_id": float64(456), "session_id": "abc"}, event.Payload)
}

func TestEvent_EdgeCases(t *testing.T) {
	t.Run("empty event type", func(t *testing.T) {
		event := NewEvent("", "value")
		assert.Equal(t, "", event.EventType)
		assert.Equal(t, gorm_types.InterfaceMap{"value": "value"}, event.Payload)
	})

	t.Run("complex payload", func(t *testing.T) {
		event := Event{EventType: "test"}
		input := map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"array": []string{"a", "b"},
		}
		err := event.SetValue(input)
		assert.NoError(t, err)
		assert.Equal(t, gorm_types.InterfaceMap{
			"nested": map[string]interface{}{"key": "value"},
			"array":  []interface{}{"a", "b"},
		}, event.Payload)
	})
}

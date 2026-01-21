// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTelemetryInterface verifies all types implement the Telemetry interface
func TestTelemetryInterface(t *testing.T) {
	tests := []struct {
		name         string
		telemetry    Telemetry
		expectedType string
	}{
		{
			name:         "Event implements Telemetry",
			telemetry:    &Event{EventType: "test_event"},
			expectedType: "event",
		},
		{
			name:         "Metric implements Telemetry",
			telemetry:    &Metric{Name: "test_metric", Value: "100"},
			expectedType: "metric",
		},
		{
			name:         "Metadata implements Telemetry",
			telemetry:    &Metadata{Key: "test_key", Value: "test_value"},
			expectedType: "metadata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedType, tt.telemetry.Type())
		})
	}
}

// TestGetDifferentTelemetry_EmptySlice tests with empty input
func TestGetDifferentTelemetry_EmptySlice(t *testing.T) {
	events, metrics, metadata := GetDifferentTelemetry([]Telemetry{})

	assert.Empty(t, events, "events should be empty")
	assert.Empty(t, metrics, "metrics should be empty")
	assert.Empty(t, metadata, "metadata should be empty")
}

// TestGetDifferentTelemetry_NilSlice tests with nil input
func TestGetDifferentTelemetry_NilSlice(t *testing.T) {
	events, metrics, metadata := GetDifferentTelemetry(nil)

	assert.NotNil(t, events, "events slice should not be nil")
	assert.NotNil(t, metrics, "metrics slice should not be nil")
	assert.NotNil(t, metadata, "metadata slice should not be nil")
	assert.Empty(t, events)
	assert.Empty(t, metrics)
	assert.Empty(t, metadata)
}

// TestGetDifferentTelemetry_SingleType tests separation with single type
func TestGetDifferentTelemetry_SingleType(t *testing.T) {
	tests := []struct {
		name              string
		input             []Telemetry
		expectedEvents    int
		expectedMetrics   int
		expectedMetadata  int
	}{
		{
			name: "Only events",
			input: []Telemetry{
				NewEvent("event1", "data1"),
				NewEvent("event2", "data2"),
				NewEvent("event3", "data3"),
			},
			expectedEvents:   3,
			expectedMetrics:  0,
			expectedMetadata: 0,
		},
		{
			name: "Only metrics",
			input: []Telemetry{
				NewMetric("metric1", "100", nil),
				NewMetric("metric2", "200", nil),
			},
			expectedEvents:   0,
			expectedMetrics:  2,
			expectedMetadata: 0,
		},
		{
			name: "Only metadata",
			input: []Telemetry{
				NewMetadata("key1", "value1"),
				NewMetadata("key2", "value2"),
				NewMetadata("key3", "value3"),
				NewMetadata("key4", "value4"),
			},
			expectedEvents:   0,
			expectedMetrics:  0,
			expectedMetadata: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events, metrics, metadata := GetDifferentTelemetry(tt.input)

			assert.Len(t, events, tt.expectedEvents, "events count mismatch")
			assert.Len(t, metrics, tt.expectedMetrics, "metrics count mismatch")
			assert.Len(t, metadata, tt.expectedMetadata, "metadata count mismatch")
		})
	}
}

// TestGetDifferentTelemetry_MixedTypes tests separation with mixed types
func TestGetDifferentTelemetry_MixedTypes(t *testing.T) {
	input := []Telemetry{
		NewEvent("call_initiated", map[string]interface{}{"call_id": "123"}),
		NewMetric("STATUS", "SUCCESS", nil),
		NewMetadata("telephony.provider", "twilio"),
		NewEvent("call_ringing", map[string]interface{}{"duration": 5}),
		NewMetadata("telephony.uuid", "sid-12345"),
		NewMetric("duration", "120", nil),
		NewEvent("call_completed", nil),
		NewMetadata("telephony.toPhone", "+1234567890"),
	}

	events, metrics, metadata := GetDifferentTelemetry(input)

	// Verify counts
	assert.Len(t, events, 3, "should have 3 events")
	assert.Len(t, metrics, 2, "should have 2 metrics")
	assert.Len(t, metadata, 3, "should have 3 metadata")

	// Verify event types
	assert.Equal(t, "call_initiated", events[0].EventType)
	assert.Equal(t, "call_ringing", events[1].EventType)
	assert.Equal(t, "call_completed", events[2].EventType)

	// Verify metric names
	assert.Equal(t, "STATUS", metrics[0].Name)
	assert.Equal(t, "SUCCESS", metrics[0].Value)
	assert.Equal(t, "duration", metrics[1].Name)
	assert.Equal(t, "120", metrics[1].Value)

	// Verify metadata keys
	assert.Equal(t, "telephony.provider", metadata[0].Key)
	assert.Equal(t, "twilio", metadata[0].Value)
	assert.Equal(t, "telephony.uuid", metadata[1].Key)
	assert.Equal(t, "telephony.toPhone", metadata[2].Key)
}

// TestGetDifferentTelemetry_OrderPreservation tests that order is preserved within each type
func TestGetDifferentTelemetry_OrderPreservation(t *testing.T) {
	input := []Telemetry{
		NewEvent("first_event", nil),
		NewMetric("first_metric", "1", nil),
		NewMetadata("first_key", "first_value"),
		NewEvent("second_event", nil),
		NewMetric("second_metric", "2", nil),
		NewMetadata("second_key", "second_value"),
		NewEvent("third_event", nil),
	}

	events, metrics, metadata := GetDifferentTelemetry(input)

	// Verify order is preserved for events
	require.Len(t, events, 3)
	assert.Equal(t, "first_event", events[0].EventType)
	assert.Equal(t, "second_event", events[1].EventType)
	assert.Equal(t, "third_event", events[2].EventType)

	// Verify order is preserved for metrics
	require.Len(t, metrics, 2)
	assert.Equal(t, "first_metric", metrics[0].Name)
	assert.Equal(t, "second_metric", metrics[1].Name)

	// Verify order is preserved for metadata
	require.Len(t, metadata, 2)
	assert.Equal(t, "first_key", metadata[0].Key)
	assert.Equal(t, "second_key", metadata[1].Key)
}

// TestGetDifferentTelemetry_RealWorldScenario tests a realistic telephony scenario
func TestGetDifferentTelemetry_RealWorldScenario(t *testing.T) {
	// Simulating what Twilio provider might return
	twilioTelemetry := []Telemetry{
		NewMetadata("telephony.provider", "twilio"),
		NewMetadata("telephony.toPhone", "+1234567890"),
		NewMetadata("telephony.fromPhone", "+0987654321"),
		NewMetadata("telephony.uuid", "CA1234567890abcdef"),
		NewEvent("initiated", map[string]interface{}{
			"AccountSid": "AC123",
			"CallSid":    "CA1234567890abcdef",
		}),
		NewMetric("STATUS", "SUCCESS", nil),
	}

	events, metrics, metadata := GetDifferentTelemetry(twilioTelemetry)

	// Verify metadata
	require.Len(t, metadata, 4)
	assert.Equal(t, "telephony.provider", metadata[0].Key)
	assert.Equal(t, "twilio", metadata[0].Value)
	assert.Equal(t, "telephony.uuid", metadata[3].Key)

	// Verify events
	require.Len(t, events, 1)
	assert.Equal(t, "initiated", events[0].EventType)
	assert.NotNil(t, events[0].Payload)

	// Verify metrics
	require.Len(t, metrics, 1)
	assert.Equal(t, "STATUS", metrics[0].Name)
	assert.Equal(t, "SUCCESS", metrics[0].Value)
}

// TestGetDifferentTelemetry_ExotelScenario tests Exotel provider scenario
func TestGetDifferentTelemetry_ExotelScenario(t *testing.T) {
	exotelTelemetry := []Telemetry{
		NewMetadata("telephony.toPhone", "+1234567890"),
		NewMetadata("telephony.fromPhone", "+0987654321"),
		NewMetadata("telephony.provider", "exotel"),
		NewMetadata("telephony.uuid", "exotel-call-123"),
		NewEvent("in-progress", map[string]interface{}{
			"Status":      "in-progress",
			"Sid":         "exotel-call-123",
			"DateCreated": "2026-01-21T10:00:00Z",
		}),
		NewMetric("STATUS", "SUCCESS", nil),
	}

	events, metrics, metadata := GetDifferentTelemetry(exotelTelemetry)

	assert.Len(t, metadata, 4)
	assert.Len(t, events, 1)
	assert.Len(t, metrics, 1)

	// Verify Exotel-specific data
	assert.Equal(t, "exotel", metadata[2].Value)
	assert.Equal(t, "in-progress", events[0].EventType)
}

// TestGetDifferentTelemetry_VonageScenario tests Vonage provider scenario
func TestGetDifferentTelemetry_VonageScenario(t *testing.T) {
	vonageTelemetry := []Telemetry{
		NewMetadata("telephony.toPhone", "+1234567890"),
		NewMetadata("telephony.fromPhone", "+0987654321"),
		NewMetadata("telephony.provider", "vonage"),
		NewMetadata("telephony.conversation_uuid", "CON-abc123"),
		NewMetadata("telephony.uuid", "vonage-uuid-456"),
		NewEvent("started", map[string]interface{}{
			"uuid":              "vonage-uuid-456",
			"conversation_uuid": "CON-abc123",
			"status":            "started",
		}),
		NewMetric("STATUS", "SUCCESS", nil),
	}

	events, metrics, metadata := GetDifferentTelemetry(vonageTelemetry)

	assert.Len(t, metadata, 5)
	assert.Len(t, events, 1)
	assert.Len(t, metrics, 1)

	// Verify Vonage has conversation_uuid
	hasConversationUUID := false
	for _, m := range metadata {
		if m.Key == "telephony.conversation_uuid" {
			hasConversationUUID = true
			assert.Equal(t, "CON-abc123", m.Value)
		}
	}
	assert.True(t, hasConversationUUID, "should have conversation_uuid metadata")
}

// TestGetDifferentTelemetry_ErrorScenario tests error telemetry
func TestGetDifferentTelemetry_ErrorScenario(t *testing.T) {
	errorTelemetry := []Telemetry{
		NewMetadata("telephony.provider", "twilio"),
		NewMetadata("telephony.toPhone", "+1234567890"),
		NewMetadata("telephony.error", "authentication failed"),
		NewEvent("FAILED", map[string]interface{}{
			"error_code":    "20003",
			"error_message": "Authentication Error",
		}),
		NewMetric("STATUS", "FAILED", nil),
	}

	events, metrics, metadata := GetDifferentTelemetry(errorTelemetry)

	assert.Len(t, metadata, 3)
	assert.Len(t, events, 1)
	assert.Len(t, metrics, 1)

	// Verify error event
	assert.Equal(t, "FAILED", events[0].EventType)
	assert.NotNil(t, events[0].Payload["error_code"])

	// Verify error metric
	assert.Equal(t, "FAILED", metrics[0].Value)
}

// TestGetDifferentTelemetry_LargeVolume tests performance with large number of telemetries
func TestGetDifferentTelemetry_LargeVolume(t *testing.T) {
	// Create 1000 telemetries (mix of all types)
	input := make([]Telemetry, 0, 1000)
	for i := 0; i < 1000; i++ {
		switch i % 3 {
		case 0:
			input = append(input, NewEvent("event", map[string]interface{}{"index": i}))
		case 1:
			input = append(input, NewMetric("metric", "value", nil))
		case 2:
			input = append(input, NewMetadata("key", "value"))
		}
	}

	events, metrics, metadata := GetDifferentTelemetry(input)

	// Should have roughly 333-334 of each
	assert.InDelta(t, 334, len(events), 1, "events count")
	assert.InDelta(t, 333, len(metrics), 1, "metrics count")
	assert.InDelta(t, 333, len(metadata), 1, "metadata count")
}

// BenchmarkGetDifferentTelemetry benchmarks the separation function
func BenchmarkGetDifferentTelemetry(b *testing.B) {
	// Create mixed telemetry slice
	input := []Telemetry{
		NewEvent("event1", "data"),
		NewMetric("metric1", "100", nil),
		NewMetadata("key1", "value1"),
		NewEvent("event2", "data"),
		NewMetric("metric2", "200", nil),
		NewMetadata("key2", "value2"),
		NewEvent("event3", "data"),
		NewMetric("metric3", "300", nil),
		NewMetadata("key3", "value3"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDifferentTelemetry(input)
	}
}

// BenchmarkGetDifferentTelemetry_LargeVolume benchmarks with 1000 items
func BenchmarkGetDifferentTelemetry_LargeVolume(b *testing.B) {
	input := make([]Telemetry, 1000)
	for i := 0; i < 1000; i++ {
		switch i % 3 {
		case 0:
			input[i] = NewEvent("event", "data")
		case 1:
			input[i] = NewMetric("metric", "value", nil)
		case 2:
			input[i] = NewMetadata("key", "value")
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetDifferentTelemetry(input)
	}
}

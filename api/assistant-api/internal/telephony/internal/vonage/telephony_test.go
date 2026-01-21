// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_vonage_telephony

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rapidaai/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//  GET /v1/talk/vonage/call/2263072539095859200?x-api-key=3dd5c2eef53d27942bccd892750fda23ea0b92965d4699e73d8e754ab882955f&to=12019868532&endpoint_type=phone&from=15703768754&conversation_uuid=CON-3d4ae1dd-5e14-4131-be3d-0247cb19a28a&uuid=bccbc3faaf864e1641fe0cdb1921b6aa&region_url=https%3A%2F%2Fapi-ap-3.vonage.com [status:200 request:61ms]

// TestReceiveCall tests the ReceiveCall method with Vonage webhook parameters
func TestReceiveCall(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedError  bool
		expectedPhone  string
		checkTelemetry func(*testing.T, []types.Telemetry)
	}{
		{
			name: "Valid Vonage webhook with all parameters",
			queryParams: map[string]string{
				"from":              "15703768754",
				"to":                "12019868532",
				"endpoint_type":     "phone",
				"conversation_uuid": "CON-3d4ae1dd-5e14-4131-be3d-0247cb19a28a",
				"uuid":              "bccbc3faaf864e1641fe0cdb1921b6aa",
				"region_url":        "https://api-ap-3.vonage.com",
			},
			expectedError: false,
			expectedPhone: "15703768754",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				require.NotNil(t, telemetry)
				assert.GreaterOrEqual(t, len(telemetry), 3)

				// Check for conversation_uuid metadata
				foundConvUUID := false
				foundUUID := false
				foundEvent := false
				foundMetric := false

				for _, tel := range telemetry {
					if metadata, ok := tel.(*types.Metadata); ok {
						if metadata.Key == "telephony.conversation_uuid" {
							assert.Equal(t, "CON-3d4ae1dd-5e14-4131-be3d-0247cb19a28a", metadata.Value)
							foundConvUUID = true
						}
						if metadata.Key == "telephony.uuid" {
							assert.Equal(t, "bccbc3faaf864e1641fe0cdb1921b6aa", metadata.Value)
							foundUUID = true
						}
					}
					if event, ok := tel.(*types.Event); ok {
						if event.EventType == "webhook" {
							foundEvent = true
							assert.NotNil(t, event.Payload)
						}
					}
					if metric, ok := tel.(*types.Metric); ok {
						if metric.Name == "STATUS" {
							assert.Equal(t, "SUCCESS", metric.Value)
							foundMetric = true
						}
					}
				}

				assert.True(t, foundConvUUID, "Should have conversation_uuid metadata")
				assert.True(t, foundUUID, "Should have uuid metadata")
				assert.True(t, foundEvent, "Should have webhook event")
				assert.True(t, foundMetric, "Should have STATUS metric")
			},
		},
		{
			name: "Valid webhook with minimal parameters",
			queryParams: map[string]string{
				"from": "15703768754",
				"to":   "12019868532",
			},
			expectedError: false,
			expectedPhone: "15703768754",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				require.NotNil(t, telemetry)

				// Should still have event and metric even without optional params
				foundEvent := false
				foundMetric := false

				for _, tel := range telemetry {
					if event, ok := tel.(*types.Event); ok {
						if event.EventType == "webhook" {
							foundEvent = true
						}
					}
					if metric, ok := tel.(*types.Metric); ok {
						if metric.Name == "STATUS" {
							foundMetric = true
						}
					}
				}

				assert.True(t, foundEvent, "Should have webhook event")
				assert.True(t, foundMetric, "Should have STATUS metric")
			},
		},
		{
			name: "Missing 'from' parameter",
			queryParams: map[string]string{
				"to":                "12019868532",
				"conversation_uuid": "CON-3d4ae1dd-5e14-4131-be3d-0247cb19a28a",
			},
			expectedError: true,
			expectedPhone: "",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				// Telemetry may be empty or contain partial data
			},
		},
		{
			name: "Empty 'from' parameter",
			queryParams: map[string]string{
				"from": "",
				"to":   "12019868532",
			},
			expectedError: true,
			expectedPhone: "",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				// Telemetry may be empty or contain partial data
			},
		},
		{
			name: "Only conversation_uuid without phone",
			queryParams: map[string]string{
				"conversation_uuid": "CON-3d4ae1dd-5e14-4131-be3d-0247cb19a28a",
				"uuid":              "bccbc3faaf864e1641fe0cdb1921b6aa",
			},
			expectedError: true,
			expectedPhone: "",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				// Telemetry may be empty or contain partial data
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Build query string
			queryValues := url.Values{}
			for key, value := range tt.queryParams {
				queryValues.Add(key, value)
			}

			// Create request with query parameters
			req := httptest.NewRequest(http.MethodGet, "/?"+queryValues.Encode(), nil)
			c.Request = req

			// Create telephony instance
			telephony := &vonageTelephony{}

			// Call ReceiveCall
			clientNumber, telemetry, err := telephony.ReceiveCall(c)

			// Verify error expectation
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, clientNumber)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, clientNumber)
				assert.Equal(t, tt.expectedPhone, *clientNumber)
			}

			// Check telemetry
			if tt.checkTelemetry != nil {
				tt.checkTelemetry(t, telemetry)
			}
		})
	}
}

// TestReceiveCall_QueryParameterExtraction tests that all query parameters are captured in telemetry
func TestReceiveCall_QueryParameterExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	queryParams := map[string]string{
		"from":              "15703768754",
		"to":                "12019868532",
		"endpoint_type":     "phone",
		"conversation_uuid": "CON-3d4ae1dd-5e14-4131-be3d-0247cb19a28a",
		"uuid":              "bccbc3faaf864e1641fe0cdb1921b6aa",
		"region_url":        "https://api-ap-3.vonage.com",
		"x-api-key":         "3dd5c2eef53d27942bccd892750fda23ea0b92965d4699e73d8e754ab882955f",
	}

	queryValues := url.Values{}
	for key, value := range queryParams {
		queryValues.Add(key, value)
	}

	req := httptest.NewRequest(http.MethodGet, "/?"+queryValues.Encode(), nil)
	c.Request = req

	telephony := &vonageTelephony{}
	_, telemetry, err := telephony.ReceiveCall(c)

	require.NoError(t, err)
	require.NotNil(t, telemetry)

	// Verify event contains all query parameters
	var webhookEvent *types.Event
	for _, tel := range telemetry {
		if event, ok := tel.(*types.Event); ok && event.EventType == "webhook" {
			webhookEvent = event
			break
		}
	}

	require.NotNil(t, webhookEvent, "Should have webhook event")
	require.NotNil(t, webhookEvent.Payload, "Event should have payload")

	// Verify all query params are in the event payload
	payloadMap := webhookEvent.Payload
	require.NotNil(t, payloadMap, "Payload should not be nil")

	for key, expectedValue := range queryParams {
		actualValue, exists := payloadMap[key]
		assert.True(t, exists, "Query param '%s' should be in payload", key)
		assert.Equal(t, expectedValue, actualValue, "Value for '%s' should match", key)
	}
}

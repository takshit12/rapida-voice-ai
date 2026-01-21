// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_twilio_telephony

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

//
// /v1/talk/twilio/call/2263072539095859200?x-api-key=3dd5c2eef53d27942bccd892750fda23ea0b92965d4699e73d8e754ab882955f&Called=%2B13345895552&ToState=AL&CallerCountry=US&Direction=inbound&CallerState=PA&ToZip=36303&CallSid=CAf64ab88f90f35581dcb16e60f875ea4a&To=%2B13345895552&CallerZip=16901&ToCountry=US&StirVerstat=TN-Validation-Passed-B&CalledZip=36303&ApiVersion=2010-04-01&CalledCity=DOTHAN&CallStatus=ringing&From=%2B15703768754&AccountSid=546789087657890876DFGHJKASHDFBJK&CalledCountry=US&CallerCity=MIDDLEBURY+CENTER&ToCity=DOTHAN&FromCountry=US&Caller=%2B15703768754&FromCity=MIDDLEBURY+CENTER&CalledState=AL&FromZip=16901&FromState=PA [status:400 request:3ms]

// TestReceiveCall tests the ReceiveCall method with Twilio webhook parameters
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
			name: "Valid Twilio webhook with all parameters",
			queryParams: map[string]string{
				"Called":        "+13345895552",
				"ToState":       "AL",
				"CallerCountry": "US",
				"Direction":     "inbound",
				"CallerState":   "PA",
				"ToZip":         "36303",
				"CallSid":       "CAf64ab88f90f35581dcb16e60f875ea4a",
				"To":            "+13345895552",
				"CallerZip":     "16901",
				"ToCountry":     "US",
				"StirVerstat":   "TN-Validation-Passed-B",
				"CalledZip":     "36303",
				"ApiVersion":    "2010-04-01",
				"CalledCity":    "DOTHAN",
				"CallStatus":    "ringing",
				"From":          "+15703768754",
				"AccountSid":    "546789087657890876DFGHJKASHDFBJK",
				"CalledCountry": "US",
				"CallerCity":    "MIDDLEBURY CENTER",
				"ToCity":        "DOTHAN",
				"FromCountry":   "US",
				"Caller":        "+15703768754",
				"FromCity":      "MIDDLEBURY CENTER",
				"CalledState":   "AL",
				"FromZip":       "16901",
				"FromState":     "PA",
			},
			expectedError: false,
			expectedPhone: "+15703768754",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				require.NotNil(t, telemetry)
				assert.GreaterOrEqual(t, len(telemetry), 2)

				// Check for CallSid metadata (telephony.uuid)
				foundCallSid := false
				foundEvent := false
				foundMetric := false

				for _, tel := range telemetry {
					if metadata, ok := tel.(*types.Metadata); ok {
						if metadata.Key == "telephony.uuid" {
							assert.Equal(t, "CAf64ab88f90f35581dcb16e60f875ea4a", metadata.Value)
							foundCallSid = true
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

				assert.True(t, foundCallSid, "Should have CallSid (telephony.uuid) metadata")
				assert.True(t, foundEvent, "Should have webhook event")
				assert.True(t, foundMetric, "Should have STATUS metric")
			},
		},
		{
			name: "Valid webhook with minimal parameters",
			queryParams: map[string]string{
				"From":    "+15703768754",
				"To":      "+13345895552",
				"CallSid": "CAf64ab88f90f35581dcb16e60f875ea4a",
			},
			expectedError: false,
			expectedPhone: "+15703768754",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				require.NotNil(t, telemetry)

				// Should still have event and metric even without optional params
				foundEvent := false
				foundMetric := false
				foundCallSid := false

				for _, tel := range telemetry {
					if metadata, ok := tel.(*types.Metadata); ok {
						if metadata.Key == "telephony.uuid" {
							foundCallSid = true
						}
					}
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

				assert.True(t, foundCallSid, "Should have CallSid metadata")
				assert.True(t, foundEvent, "Should have webhook event")
				assert.True(t, foundMetric, "Should have STATUS metric")
			},
		},
		{
			name: "Valid webhook without CallSid",
			queryParams: map[string]string{
				"From": "+15703768754",
				"To":   "+13345895552",
			},
			expectedError: false,
			expectedPhone: "+15703768754",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				require.NotNil(t, telemetry)

				// Should have event and metric but no CallSid metadata
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
			name: "Missing 'From' parameter",
			queryParams: map[string]string{
				"To":         "+13345895552",
				"CallSid":    "CAf64ab88f90f35581dcb16e60f875ea4a",
				"CallStatus": "ringing",
			},
			expectedError: true,
			expectedPhone: "",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				// Telemetry may be empty or contain partial data
			},
		},
		{
			name: "Empty 'From' parameter",
			queryParams: map[string]string{
				"From": "",
				"To":   "+13345895552",
			},
			expectedError: true,
			expectedPhone: "",
			checkTelemetry: func(t *testing.T, telemetry []types.Telemetry) {
				// Telemetry may be empty or contain partial data
			},
		},
		{
			name: "Only CallSid without From",
			queryParams: map[string]string{
				"CallSid":    "CAf64ab88f90f35581dcb16e60f875ea4a",
				"To":         "+13345895552",
				"CallStatus": "ringing",
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
			telephony := &twilioTelephony{}

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
		"Called":        "+13345895552",
		"ToState":       "AL",
		"CallerCountry": "US",
		"Direction":     "inbound",
		"CallerState":   "PA",
		"ToZip":         "36303",
		"CallSid":       "CAf64ab88f90f35581dcb16e60f875ea4a",
		"To":            "+13345895552",
		"CallerZip":     "16901",
		"ToCountry":     "US",
		"StirVerstat":   "TN-Validation-Passed-B",
		"CalledZip":     "36303",
		"ApiVersion":    "2010-04-01",
		"CalledCity":    "DOTHAN",
		"CallStatus":    "ringing",
		"From":          "+15703768754",
		"AccountSid":    "546789087657890876DFGHJKASHDFBJK",
		"CalledCountry": "US",
		"CallerCity":    "MIDDLEBURY CENTER",
		"ToCity":        "DOTHAN",
		"FromCountry":   "US",
		"Caller":        "+15703768754",
		"FromCity":      "MIDDLEBURY CENTER",
		"CalledState":   "AL",
		"FromZip":       "16901",
		"FromState":     "PA",
	}

	queryValues := url.Values{}
	for key, value := range queryParams {
		queryValues.Add(key, value)
	}

	req := httptest.NewRequest(http.MethodGet, "/?"+queryValues.Encode(), nil)
	c.Request = req

	telephony := &twilioTelephony{}
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

// TestReceiveCall_PhoneNumberFormats tests various phone number formats
func TestReceiveCall_PhoneNumberFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		phoneNumber   string
		expectedPhone string
	}{
		{
			name:          "E.164 format with plus",
			phoneNumber:   "+15703768754",
			expectedPhone: "+15703768754",
		},
		{
			name:          "10-digit US number",
			phoneNumber:   "5703768754",
			expectedPhone: "5703768754",
		},
		{
			name:          "International format",
			phoneNumber:   "+441234567890",
			expectedPhone: "+441234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			queryValues := url.Values{}
			queryValues.Add("From", tt.phoneNumber)
			queryValues.Add("To", "+13345895552")

			req := httptest.NewRequest(http.MethodGet, "/?"+queryValues.Encode(), nil)
			c.Request = req

			telephony := &twilioTelephony{}
			clientNumber, _, err := telephony.ReceiveCall(c)

			require.NoError(t, err)
			require.NotNil(t, clientNumber)
			assert.Equal(t, tt.expectedPhone, *clientNumber)
		})
	}
}

// TestReceiveCall_TelemetryStructure tests the structure of telemetry data
func TestReceiveCall_TelemetryStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	queryValues := url.Values{}
	queryValues.Add("From", "+15703768754")
	queryValues.Add("To", "+13345895552")
	queryValues.Add("CallSid", "CAf64ab88f90f35581dcb16e60f875ea4a")
	queryValues.Add("CallStatus", "ringing")

	req := httptest.NewRequest(http.MethodGet, "/?"+queryValues.Encode(), nil)
	c.Request = req

	telephony := &twilioTelephony{}
	_, telemetry, err := telephony.ReceiveCall(c)

	require.NoError(t, err)
	require.NotNil(t, telemetry)

	// Count different types of telemetry
	metadataCount := 0
	eventCount := 0
	metricCount := 0

	for _, tel := range telemetry {
		switch tel.(type) {
		case *types.Metadata:
			metadataCount++
		case *types.Event:
			eventCount++
		case *types.Metric:
			metricCount++
		}
	}

	// Should have exactly 1 metadata (CallSid), 1 event (webhook), and 1 metric (STATUS)
	assert.Equal(t, 1, metadataCount, "Should have exactly 1 metadata entry")
	assert.Equal(t, 1, eventCount, "Should have exactly 1 event entry")
	assert.Equal(t, 1, metricCount, "Should have exactly 1 metric entry")
}

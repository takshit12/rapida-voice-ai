// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package endpoint_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/rapidaai/pkg/clients/rest"
	"github.com/rapidaai/pkg/commons"
)

const (
	calcomBaseURL             = "https://api.cal.com"
	calcomSlotsAPIVersion     = "2024-09-04"
	calcomBookingsAPIVersion  = "2024-08-13"
)

// CalcomHandler handles Cal.com availability checks and booking creation.
//
// Configuration:
//   - Cal.com API token: CALCOM_API_TOKEN environment variable
//
// This handler exposes two endpoints:
//   - POST /v1/calcom/availability  — returns available time slots for a given date range
//   - POST /v1/calcom/book          — creates a booking on Cal.com
//
// When configuring api_request tools in the UI, use one of:
//   - Via gateway: https://<gateway-host>/api/endpoint/v1/calcom/availability and /v1/calcom/book
//   - Direct:      https://<endpoint-api-host>/v1/calcom/availability and /v1/calcom/book
type CalcomHandler struct {
	logger   commons.Logger
	apiToken string
	client   *rest.RestClient
}

// NewCalcomHandler creates a new Cal.com handler instance.
// Reads CALCOM_API_TOKEN from the environment at startup.
func NewCalcomHandler(logger commons.Logger) *CalcomHandler {
	token := os.Getenv("CALCOM_API_TOKEN")
	client := rest.NewRestClientWithConfig(calcomBaseURL, map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}, 30)

	return &CalcomHandler{
		logger:   logger,
		apiToken: token,
		client:   client,
	}
}

// --------------------------------------------------------------------------
// Availability Handler
// --------------------------------------------------------------------------

// availabilityRequest is the expected JSON body for the availability endpoint.
type availabilityRequest struct {
	EventTypeID int    `json:"eventTypeId" binding:"required"`
	StartDate   string `json:"startDate" binding:"required"` // ISO 8601, e.g. "2026-02-15T00:00:00.000Z"
	EndDate     string `json:"endDate" binding:"required"`   // ISO 8601, e.g. "2026-02-22T23:59:59.000Z"
}

// availabilityRequestWrapper wraps the availability request.
// The api_request tool caller sends the LLM's function-call args under a
// single key (configured via tool.argument → "data" in the UI), so the
// incoming POST body looks like: { "data": { "eventTypeId":…, … } }.
type availabilityRequestWrapper struct {
	Data availabilityRequest `json:"data" binding:"required"`
}

// HandleAvailability queries Cal.com v2 /slots endpoint and returns available time slots.
//
// Expected body from the api_request tool caller (JSON):
//
//	{
//	  "data": {
//	    "eventTypeId": 4704916,
//	    "startDate": "2026-02-15T00:00:00.000Z",
//	    "endDate": "2026-02-22T23:59:59.000Z"
//	  }
//	}
//
// Returns the available slots grouped by date.
func (h *CalcomHandler) HandleAvailability(c *gin.Context) {
	if h.apiToken == "" {
		h.logger.Error("CALCOM_API_TOKEN environment variable is not set")
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Cal.com integration is not configured on the server",
		})
		return
	}

	var wrapper availabilityRequestWrapper
	if err := c.ShouldBindJSON(&wrapper); err != nil {
		h.logger.Errorf("Failed to parse availability request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}
	req := wrapper.Data

	// Build query parameters for Cal.com v2 GET /slots
	params := map[string]interface{}{
		"eventTypeId": req.EventTypeID,
		"start":       req.StartDate,
		"end":         req.EndDate,
	}

	headers := map[string]string{
		"cal-api-version": calcomSlotsAPIVersion,
	}

	ctx := context.Background()
	resp, err := h.client.Get(ctx, "/v2/slots", params, headers)
	if err != nil {
		h.logger.Errorf("Failed to call Cal.com slots API: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Failed to reach Cal.com API: %v", err),
		})
		return
	}

	// Parse the Cal.com response
	var calcomResp map[string]interface{}
	if err := json.Unmarshal(resp.Body, &calcomResp); err != nil {
		h.logger.Errorf("Failed to parse Cal.com slots response: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"ok":    false,
			"error": "Failed to parse Cal.com response",
		})
		return
	}

	// Check if Cal.com returned an error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.logger.Errorf("Cal.com slots API returned status %d: %s", resp.StatusCode, resp.ToString())
		c.JSON(resp.StatusCode, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Cal.com API error (status %d)", resp.StatusCode),
			"data":  calcomResp,
		})
		return
	}

	h.logger.Infof("Successfully fetched availability for event type %d", req.EventTypeID)
	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"data": calcomResp["data"],
	})
}

// --------------------------------------------------------------------------
// Booking Handler
// --------------------------------------------------------------------------

// bookingAttendee represents the attendee information for a booking.
type bookingAttendee struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	TimeZone string `json:"timeZone" binding:"required"` // IANA timezone, e.g. "Asia/Kolkata"
	Language string `json:"language,omitempty"`           // BCP 47 language code, e.g. "en"
}

// bookingRequest is the expected JSON body for the booking endpoint.
type bookingRequest struct {
	EventTypeID int                    `json:"eventTypeId" binding:"required"`
	Start       string                 `json:"start" binding:"required"` // ISO 8601 slot time from availability response
	Attendee    bookingAttendee        `json:"attendee" binding:"required"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
}

// bookingRequestWrapper wraps the booking request.
// The api_request tool caller sends the LLM's function-call args under a
// single key (configured via tool.argument → "data" in the UI), so the
// incoming POST body looks like: { "data": { "eventTypeId":…, … } }.
type bookingRequestWrapper struct {
	Data bookingRequest `json:"data" binding:"required"`
}

// HandleBook creates a booking on Cal.com v2 /bookings endpoint.
//
// Expected body from the api_request tool caller (JSON):
//
//	{
//	  "data": {
//	    "eventTypeId": 4704916,
//	    "start": "2026-02-17T09:00:00.000Z",
//	    "attendee": {
//	      "name": "John Doe",
//	      "email": "john@example.com",
//	      "timeZone": "Asia/Kolkata",
//	      "language": "en"
//	    },
//	    "metadata": { "source": "voice-ai" },
//	    "notes": "Booked via Rapida Voice AI"
//	  }
//	}
//
// Returns the created booking details including confirmation.
func (h *CalcomHandler) HandleBook(c *gin.Context) {
	if h.apiToken == "" {
		h.logger.Error("CALCOM_API_TOKEN environment variable is not set")
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Cal.com integration is not configured on the server",
		})
		return
	}

	var wrapper bookingRequestWrapper
	if err := c.ShouldBindJSON(&wrapper); err != nil {
		h.logger.Errorf("Failed to parse booking request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}
	req := wrapper.Data

	// Cal.com requires attendee.language to be one of its allowed locale codes; default to "en" if omitted
	attendeeLang := req.Attendee.Language
	if attendeeLang == "" {
		attendeeLang = "en"
	}

	// Build the Cal.com v2 booking payload
	calcomPayload := map[string]interface{}{
		"eventTypeId": req.EventTypeID,
		"start":       req.Start,
		"attendee": map[string]interface{}{
			"name":     req.Attendee.Name,
			"email":    req.Attendee.Email,
			"timeZone": req.Attendee.TimeZone,
			"language": attendeeLang,
		},
	}

	// Add optional metadata
	if req.Metadata != nil {
		calcomPayload["metadata"] = req.Metadata
	}

	// Add notes as bookingFieldsResponses if provided
	if req.Notes != "" {
		calcomPayload["bookingFieldsResponses"] = map[string]interface{}{
			"notes": req.Notes,
		}
	}

	headers := map[string]string{
		"cal-api-version": calcomBookingsAPIVersion,
	}

	ctx := context.Background()
	resp, err := h.client.Post(ctx, "/v2/bookings", calcomPayload, headers)
	if err != nil {
		h.logger.Errorf("Failed to call Cal.com bookings API: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Failed to reach Cal.com API: %v", err),
		})
		return
	}

	// Parse the Cal.com response
	var calcomResp map[string]interface{}
	if err := json.Unmarshal(resp.Body, &calcomResp); err != nil {
		h.logger.Errorf("Failed to parse Cal.com booking response: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"ok":    false,
			"error": "Failed to parse Cal.com response",
		})
		return
	}

	// Check if Cal.com returned an error
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.logger.Errorf("Cal.com bookings API returned status %d: %s", resp.StatusCode, resp.ToString())
		c.JSON(resp.StatusCode, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Cal.com API error (status %d)", resp.StatusCode),
			"data":  calcomResp,
		})
		return
	}

	h.logger.Infof("Successfully created booking for event type %d at %s", req.EventTypeID, req.Start)
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Booking created successfully",
		"data":    calcomResp["data"],
	})
}

// --------------------------------------------------------------------------
// Webhook Handler
// --------------------------------------------------------------------------

// HandleWebhook receives Cal.com webhook events (e.g. booking.created, booking.canceled).
// This can be used for reconciliation and state tracking.
//
// Expected body: Cal.com webhook payload (varies by event type).
// For now, this just logs the event and returns 200.
func (h *CalcomHandler) HandleWebhook(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.logger.Errorf("Failed to parse Cal.com webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid webhook body: %v", err),
		})
		return
	}

	// Extract the trigger/event type for logging
	triggerEvent, _ := body["triggerEvent"].(string)
	h.logger.Infof("Received Cal.com webhook event: %s", triggerEvent)

	// Log the full payload at debug level for troubleshooting
	if payloadBytes, err := json.Marshal(body); err == nil {
		h.logger.Debugf("Cal.com webhook payload: %s", string(payloadBytes))
	}

	// Acknowledge the webhook immediately
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": fmt.Sprintf("Webhook event '%s' received", triggerEvent),
	})
}

// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package endpoint_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/rapidaai/pkg/commons"
)

// GoogleSheetsWebhookHandler handles incoming webhook requests from
// the Assistant API and appends extracted analysis data to a Google Sheet.
//
// Configuration:
//   - Google credentials: GOOGLE_SERVICE_ACCOUNT_JSON environment variable
//   - Target sheet: X-Sheet-ID request header (per-agent)
//   - Column mapping: X-Sheet-Columns request header (per-agent)
//   - Sheet tab name: X-Sheet-Name request header (optional, defaults to "Sheet1")
type GoogleSheetsWebhookHandler struct {
	logger             commons.Logger
	serviceAccountJSON string
}

// NewGoogleSheetsWebhookHandler creates a new handler instance.
// Reads GOOGLE_SERVICE_ACCOUNT_JSON from the environment at startup.
func NewGoogleSheetsWebhookHandler(logger commons.Logger) *GoogleSheetsWebhookHandler {
	return &GoogleSheetsWebhookHandler{
		logger:             logger,
		serviceAccountJSON: os.Getenv("GOOGLE_SERVICE_ACCOUNT_JSON"),
	}
}

// Handle processes a webhook POST request and appends a row to Google Sheets.
//
// Expected headers:
//
//	X-Sheet-ID:      Google Spreadsheet ID (from the sheet URL)
//	X-Sheet-Columns: Comma-separated column names defining the row order
//	X-Sheet-Name:    (Optional) Sheet tab name, defaults to "Sheet1"
//
// Expected body (JSON):
//
//	{
//	  "analysis": { "callerName": "...", "callerLocation": "...", ... },
//	  "conversationId": "12345",
//	  "assistantId": "67890"
//	}
func (h *GoogleSheetsWebhookHandler) Handle(c *gin.Context) {
	// ── Step 1: Validate server-side configuration ──
	if h.serviceAccountJSON == "" {
		h.logger.Error("GOOGLE_SERVICE_ACCOUNT_JSON environment variable is not set")
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Google Sheets integration is not configured on the server",
		})
		return
	}

	// ── Step 2: Validate required request headers ──
	sheetID := c.GetHeader("X-Sheet-ID")
	if sheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Missing required header: X-Sheet-ID",
		})
		return
	}

	columnsHeader := c.GetHeader("X-Sheet-Columns")
	if columnsHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "Missing required header: X-Sheet-Columns",
		})
		return
	}

	sheetName := c.GetHeader("X-Sheet-Name")
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	// Parse column names into an ordered list
	columns := strings.Split(columnsHeader, ",")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}

	// ── Step 3: Parse the webhook request body ──
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.logger.Errorf("Failed to parse webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Invalid JSON body: %v", err),
		})
		return
	}

	// Extract the nested analysis object (may be nil if not present)
	analysis, _ := body["analysis"].(map[string]interface{})

	// Safety net: if the analysis only contains a "result" key with a raw string,
	// the LLM likely wrapped its JSON in markdown fences (```json ... ```).
	// Try to strip the fences and parse the inner JSON.
	if analysis != nil {
		if result, ok := analysis["result"].(string); ok && len(analysis) == 1 {
			cleaned := strings.TrimSpace(result)
			if strings.HasPrefix(cleaned, "```") {
				if idx := strings.Index(cleaned, "\n"); idx != -1 {
					cleaned = cleaned[idx+1:]
				}
				cleaned = strings.TrimSpace(cleaned)
				if strings.HasSuffix(cleaned, "```") {
					cleaned = cleaned[:len(cleaned)-3]
				}
				cleaned = strings.TrimSpace(cleaned)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(cleaned), &parsed); err == nil {
				analysis = parsed
			}
		}
	}

	// ── Step 4: Build the row based on column mapping ──
	row := make([]interface{}, len(columns))
	for i, col := range columns {
		switch col {
		case "timestamp":
			// Special column: server-generated UTC timestamp
			row[i] = time.Now().UTC().Format(time.RFC3339)
		default:
			// Priority 1: top-level body keys (conversationId, assistantId, etc.)
			if val, ok := body[col]; ok {
				row[i] = formatCellValue(val)
			} else if analysis != nil {
				// Priority 2: keys inside the analysis object (callerName, etc.)
				if val, ok := analysis[col]; ok {
					row[i] = formatCellValue(val)
				}
			}
		}
		// Default to empty string if no value was resolved
		if row[i] == nil {
			row[i] = ""
		}
	}

	// ── Step 5: Initialize Google Sheets client ──
	ctx := c.Request.Context()
	srv, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(h.serviceAccountJSON)))
	if err != nil {
		h.logger.Errorf("Failed to create Google Sheets service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "Failed to initialize Google Sheets client",
		})
		return
	}

	// ── Step 6: Append the row to the sheet ──
	appendRange := fmt.Sprintf("%s!A:Z", sheetName)
	_, err = srv.Spreadsheets.Values.Append(sheetID, appendRange, &sheets.ValueRange{
		Values: [][]interface{}{row},
	}).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		h.logger.Errorf("Failed to append row to Google Sheet %s: %v", sheetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": fmt.Sprintf("Failed to append row to sheet: %v", err),
		})
		return
	}

	h.logger.Infof("Successfully appended row to Google Sheet %s (%d columns)", sheetID, len(columns))
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "Row appended to Google Sheet",
	})
}

// formatCellValue converts a value into a format suitable for a Google Sheets cell.
// Primitive types (string, number, bool) pass through as-is.
// Complex types (maps, slices) are JSON-serialized into a string.
func formatCellValue(val interface{}) interface{} {
	switch v := val.(type) {
	case string, float64, bool:
		return v
	case nil:
		return ""
	default:
		// For maps, slices, or any other complex type: JSON-serialize
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	}
}

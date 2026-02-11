// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package endpoint_router

import (
	"github.com/gin-gonic/gin"

	endpoint_api "github.com/rapidaai/api/endpoint-api/api"
	"github.com/rapidaai/pkg/commons"
)

// WebhookRoutes registers HTTP routes for receiving webhook callbacks.
// These routes are served by the Gin engine (not gRPC) and are
// accessible without Rapida authentication, since they are called
// by the Assistant API's internal webhook system.
func WebhookRoutes(engine *gin.Engine, logger commons.Logger) {
	logger.Info("Webhook routes added to engine.")
	handler := endpoint_api.NewGoogleSheetsWebhookHandler(logger)

	v1 := engine.Group("/v1/webhook")
	{
		v1.POST("/google-sheets", handler.Handle)
	}

	// Cal.com webhook for booking events (booking.created, booking.canceled, etc.)
	calcomHandler := endpoint_api.NewCalcomHandler(logger)
	v1.POST("/calcom", calcomHandler.HandleWebhook)
}

// CalcomRoutes registers HTTP routes for Cal.com availability and booking.
// These are called by the Assistant API's api_request tool during conversations.
func CalcomRoutes(engine *gin.Engine, logger commons.Logger) {
	logger.Info("Cal.com routes added to engine.")
	calcomHandler := endpoint_api.NewCalcomHandler(logger)

	v1 := engine.Group("/v1/calcom")
	{
		v1.POST("/availability", calcomHandler.HandleAvailability)
		v1.POST("/book", calcomHandler.HandleBook)
	}
}

// Rapida – Open Source Voice AI Orchestration Platform
// Copyright (C) 2023-2025 Prashant Srivastav <prashant@rapida.ai>
// Licensed under a modified GPL-2.0. See the LICENSE file for details.
package internal_vertexai_callers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/auth"
	internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/protos"
	"google.golang.org/genai"
)

type VertexAi struct {
	logger     commons.Logger
	credential internal_callers.CredentialResolver
}

var (
	PROJECT_ID          = "project_id"
	SERVICE_ACCOUNT_KEY = "service_account_key"
	REGION              = "region"
)

func vertexai(logger commons.Logger, credential *protos.Credential) VertexAi {
	return VertexAi{
		logger: logger,
		credential: func() map[string]interface{} {
			return credential.GetValue().AsMap()
		},
	}
}
func (goog *VertexAi) GetClient() (*genai.Client, error) {
	ctx := context.Background()
	credentials := goog.credential()

	prj, ok := credentials[PROJECT_ID]
	if !ok {
		return nil, errors.New("unable to resolve the credential")
	}
	serviceCrd, ok := credentials[SERVICE_ACCOUNT_KEY]
	if !ok {
		return nil, errors.New("unable to resolve the credential")
	}
	region, ok := credentials[REGION]
	if !ok {
		return nil, errors.New("unable to resolve the credential")
	}
	serviceCrdJSON := []byte(serviceCrd.(string))

	var googleServiceJSON struct {
		ProjectID   string `json:"project_id"`
		PrivateKey  string `json:"private_key"`
		ClientEmail string `json:"client_email"`
		TokenURI    string `json:"token_uri"`
	}

	if err := json.Unmarshal(serviceCrdJSON, &googleServiceJSON); err != nil {
		return nil, fmt.Errorf("failed to parse service account JSON: %w", err)
	}

	tp, err := auth.New2LOTokenProvider(&auth.Options2LO{
		Email:      googleServiceJSON.ClientEmail,
		PrivateKey: []byte(googleServiceJSON.PrivateKey),
		TokenURL:   googleServiceJSON.TokenURI,
		Scopes:     []string{"https://www.googleapis.com/auth/cloud-platform"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create token provider: %w", err)
	}
	return genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  prj.(string),
		Location: region.(string),
		Credentials: auth.NewCredentials(&auth.CredentialsOptions{
			TokenProvider: tp,
			JSON:          serviceCrdJSON}),
	})

}
func (goog *VertexAi) UsageMetrics(usages *genai.GenerateContentResponseUsageMetadata) types.Metrics {
	metrics := make(types.Metrics, 0)
	if usages != nil {
		metrics = append(metrics, &types.Metric{
			Name:        type_enums.INPUT_TOKEN.String(),
			Value:       fmt.Sprintf("%d", usages.PromptTokenCount),
			Description: "Input tokens (including cached content)",
		})

		if usages.CachedContentTokenCount > 0 {
			metrics = append(metrics, &types.Metric{
				Name:        "CACHED_CONTENT_TOKEN",
				Value:       fmt.Sprintf("%d", usages.CachedContentTokenCount),
				Description: "Cached content tokens",
			})
		}

		metrics = append(metrics, &types.Metric{
			Name:        type_enums.OUTPUT_TOKEN.String(),
			Value:       fmt.Sprintf("%d", usages.CandidatesTokenCount),
			Description: "Output tokens",
		})

		if usages.ToolUsePromptTokenCount > 0 {
			metrics = append(metrics, &types.Metric{
				Name:        "TOOL_USE_PROMPT_TOKEN",
				Value:       fmt.Sprintf("%d", usages.ToolUsePromptTokenCount),
				Description: "Tool-use prompt tokens",
			})
		}

		if usages.ThoughtsTokenCount > 0 {
			metrics = append(metrics, &types.Metric{
				Name:        "THOUGHTS_TOKEN",
				Value:       fmt.Sprintf("%d", usages.ThoughtsTokenCount),
				Description: "Thoughts tokens for thinking models",
			})
		}

		metrics = append(metrics, &types.Metric{
			Name:        type_enums.TOTAL_TOKEN.String(),
			Value:       fmt.Sprintf("%d", usages.TotalTokenCount),
			Description: "Total tokens (prompt, response, and tool-use)",
		})
	}
	return metrics
}

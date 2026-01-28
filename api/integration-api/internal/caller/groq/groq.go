package internal_groq_callers

import (
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	integration_api "github.com/rapidaai/protos"
)

type Groq struct {
	logger     commons.Logger
	credential internal_callers.CredentialResolver
}

var (
	GROQ_BASE_URL = "https://api.groq.com/openai/v1"
	API_KEY       = "key"
)

const (
	// ChatRoleAssistant - The role that provides responses to system-instructed, user-prompted input.
	ChatRoleAssistant string = "assistant"
	// ChatRoleFunction - The role that provides function results for chat completions.
	ChatRoleFunction string = "function"
	// ChatRoleSystem - The role that instructs or sets the behavior of the assistant.
	ChatRoleSystem string = "system"
	// ChatRoleTool - The role that represents extension tool activity within a chat completions operation.
	ChatRoleTool string = "tool"
	// ChatRoleUser - The role that provides input for chat completions.
	ChatRoleUser string = "user"
)

func groq(logger commons.Logger, credential *integration_api.Credential) Groq {
	_credential := credential.GetValue().AsMap()
	return Groq{logger: logger,
		credential: func() map[string]interface{} {
			return _credential
		}}
}

func (g *Groq) GetClient() (*openai.Client, error) {
	g.logger.Debugf("Getting client for Groq")
	credentials := g.credential()
	apiKey, ok := credentials[API_KEY]
	if !ok {
		g.logger.Errorf("Unable to get client for Groq - missing API key")
		return nil, errors.New("unable to resolve the credential")
	}
	clt := openai.NewClient(
		option.WithAPIKey(apiKey.(string)),
		option.WithBaseURL(GROQ_BASE_URL),
	)
	return &clt, nil
}

func (g *Groq) GetComplitionUsages(usages openai.CompletionUsage) types.Metrics {
	metrics := make(types.Metrics, 0)
	metrics = append(metrics, &types.Metric{
		Name:        type_enums.OUTPUT_TOKEN.String(),
		Value:       fmt.Sprintf("%d", usages.CompletionTokens),
		Description: "Input token",
	})

	metrics = append(metrics, &types.Metric{
		Name:        type_enums.INPUT_TOKEN.String(),
		Value:       fmt.Sprintf("%d", usages.PromptTokens),
		Description: "Output Token",
	})

	metrics = append(metrics, &types.Metric{
		Name:        type_enums.TOTAL_TOKEN.String(),
		Value:       fmt.Sprintf("%d", usages.TotalTokens),
		Description: "Total Token",
	})
	return metrics
}

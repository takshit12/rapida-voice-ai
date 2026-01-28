package internal_groq_callers

import (
	"context"
	"time"

	"github.com/openai/openai-go"

	internal_callers "github.com/rapidaai/api/integration-api/internal/caller"
	"github.com/rapidaai/pkg/commons"
	integration_api "github.com/rapidaai/protos"
)

type verifyCredentialCaller struct {
	Groq
}

func NewVerifyCredentialCaller(logger commons.Logger, credential *integration_api.Credential) internal_callers.Verifier {
	return &verifyCredentialCaller{
		Groq: groq(logger, credential),
	}
}

func (stc *verifyCredentialCaller) CredentialVerifier(
	ctx context.Context,
	options *internal_callers.CredentialVerifierOptions) (*string, error) {
	client, err := stc.GetClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	_, err = client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Test"),
		},
		Model: "llama-3.3-70b-versatile",
	})
	if err != nil {
		return nil, err
	}
	return nil, err
}

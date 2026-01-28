// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package integration_client

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/rapidaai/config"
	"github.com/rapidaai/pkg/clients"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/connectors"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/protos"
)

type IntegrationServiceClient interface {
	Chat(c context.Context,
		auth types.SimplePrinciple,
		providerName string,
		request *protos.ChatRequest) (*protos.ChatResponse, error)
	StreamChat(c context.Context, auth types.SimplePrinciple,
		providerName string,
		request *protos.ChatRequest) (protos.OpenAiService_StreamChatClient, error)
	Embedding(ctx context.Context, auth types.SimplePrinciple, providerName string, in *protos.EmbeddingRequest) (*protos.EmbeddingResponse, error)
	Reranking(ctx context.Context, auth types.SimplePrinciple, providerName string, in *protos.RerankingRequest) (*protos.RerankingResponse, error)
	VerifyCredential(ctx context.Context, auth types.SimplePrinciple, providerName string, in *protos.Credential) (*protos.VerifyCredentialResponse, error)
}

type integrationServiceClient struct {
	clients.InternalClient
	cfg               *config.AppConfig
	logger            commons.Logger
	cohereClient      protos.CohereServiceClient
	replicateClient   protos.ReplicateServiceClient
	openAiClient      protos.OpenAiServiceClient
	voyageAiClient    protos.VoyageAiServiceClient
	bedrockClient     protos.BedrockServiceClient
	azureAiClient     protos.AzureServiceClient
	anthropicClient   protos.AnthropicServiceClient
	geminiClient      protos.GeminiServiceClient
	vertexaiClient    protos.VertexAiServiceClient
	mistralClient     protos.MistralServiceClient
	togetherAiClient  protos.TogetherAiServiceClient
	deepInfraCLient   protos.DeepInfraServiceClient
	huggingfaceClient protos.HuggingfaceServiceClient
	awsbedrockClient  protos.BedrockServiceClient
}

func NewIntegrationServiceClientGRPC(config *config.AppConfig, logger commons.Logger, redis connectors.RedisConnector) IntegrationServiceClient {
	lightConnection, err := grpc.NewClient(config.IntegrationHost, []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}...)
	if err != nil {
		logger.Fatalf("Unable to create connection %v", err)
	}
	return &integrationServiceClient{
		InternalClient:    clients.NewInternalClient(config, logger, redis),
		cfg:               config,
		logger:            logger,
		cohereClient:      protos.NewCohereServiceClient(lightConnection),
		replicateClient:   protos.NewReplicateServiceClient(lightConnection),
		openAiClient:      protos.NewOpenAiServiceClient(lightConnection),
		anthropicClient:   protos.NewAnthropicServiceClient(lightConnection),
		geminiClient:      protos.NewGeminiServiceClient(lightConnection),
		vertexaiClient:    protos.NewVertexAiServiceClient(lightConnection),
		mistralClient:     protos.NewMistralServiceClient(lightConnection),
		togetherAiClient:  protos.NewTogetherAiServiceClient(lightConnection),
		deepInfraCLient:   protos.NewDeepInfraServiceClient(lightConnection),
		voyageAiClient:    protos.NewVoyageAiServiceClient(lightConnection),
		bedrockClient:     protos.NewBedrockServiceClient(lightConnection),
		azureAiClient:     protos.NewAzureServiceClient(lightConnection),
		huggingfaceClient: protos.NewHuggingfaceServiceClient(lightConnection),
		awsbedrockClient:  protos.NewBedrockServiceClient(lightConnection),
	}
}

func (client *integrationServiceClient) Embedding(c context.Context,
	auth types.SimplePrinciple,
	providerName string,
	request *protos.EmbeddingRequest) (*protos.EmbeddingResponse, error) {
	switch providerName := strings.ToLower(providerName); providerName {
	case "cohere":
		return client.cohereClient.Embedding(client.WithAuth(c, auth), request)
	case "openai":
		return client.openAiClient.Embedding(client.WithAuth(c, auth), request)
	case "voyageai":
		return client.voyageAiClient.Embedding(client.WithAuth(c, auth), request)
	case "bedrock":
		return client.bedrockClient.Embedding(client.WithAuth(c, auth), request)
	case "azure-foundry":
		return client.azureAiClient.Embedding(client.WithAuth(c, auth), request)
	case "gemini":
		return client.geminiClient.Embedding(client.WithAuth(c, auth), request)
	// case "mistral":
	// return client.mistralClient.Embedding(client.WithAuth(c, auth), request)
	default:
		return nil, errors.New("illegal provider for chat request")
	}
}

func (client *integrationServiceClient) Reranking(c context.Context,
	auth types.SimplePrinciple,
	providerName string,
	request *protos.RerankingRequest) (*protos.RerankingResponse, error) {
	switch providerName := strings.ToLower(providerName); providerName {
	case "cohere":
		return client.cohereClient.Reranking(client.WithAuth(c, auth), request)
	default:
		return nil, errors.New("illegal provider for chat request")
	}
}

func (client *integrationServiceClient) Chat(c context.Context,
	auth types.SimplePrinciple,
	providerName string,
	request *protos.ChatRequest) (*protos.ChatResponse, error) {
	switch providerName := strings.ToLower(providerName); providerName {
	case "cohere":
		return client.cohereClient.Chat(client.WithAuth(c, auth), request)
	case "anthropic":
		return client.anthropicClient.Chat(client.WithAuth(c, auth), request)
	case "replicate":
		return client.replicateClient.Chat(client.WithAuth(c, auth), request)
	case "gemini":
		return client.geminiClient.Chat(client.WithAuth(c, auth), request)
	case "mistral":
		return client.mistralClient.Chat(client.WithAuth(c, auth), request)
	case "togetherai":
		return client.togetherAiClient.Chat(client.WithAuth(c, auth), request)
	case "openai":
		return client.openAiClient.Chat(client.WithAuth(c, auth), request)
	case "groq":
		// Groq uses OpenAI-compatible API, reuse OpenAI client
		return client.openAiClient.Chat(client.WithAuth(c, auth), request)
	case "aws-bedrock":
		return client.bedrockClient.Chat(client.WithAuth(c, auth), request)
	case "azure-foundry":
		return client.azureAiClient.Chat(client.WithAuth(c, auth), request)
	case "vertexai":
		return client.vertexaiClient.Chat(client.WithAuth(c, auth), request)
	default:
		return nil, errors.New("illegal provider for chat request")
	}
}

// StreamChat implements IntegrationServiceClient.
func (client *integrationServiceClient) StreamChat(c context.Context, auth types.SimplePrinciple, providerName string, request *protos.ChatRequest) (protos.OpenAiService_StreamChatClient, error) {
	switch providerName := strings.ToLower(providerName); providerName {
	case "openai":
		return client.openAiClient.StreamChat(client.WithAuth(c, auth), request)
	case "groq":
		// Groq uses OpenAI-compatible API, reuse OpenAI client
		return client.openAiClient.StreamChat(client.WithAuth(c, auth), request)
	case "anthropic":
		return client.anthropicClient.StreamChat(client.WithAuth(c, auth), request)
	case "gemini":
		return client.geminiClient.StreamChat(client.WithAuth(c, auth), request)
	case "cohere":
		return client.cohereClient.StreamChat(client.WithAuth(c, auth), request)
	case "azure-foundry":
		return client.azureAiClient.StreamChat(client.WithAuth(c, auth), request)
	case "vertexai":
		return client.vertexaiClient.StreamChat(client.WithAuth(c, auth), request)
	default:
		return nil, errors.New("illegal provider for chat request")
	}
}

func (client *integrationServiceClient) VerifyCredential(c context.Context,
	auth types.SimplePrinciple,
	providerName string,
	cr *protos.Credential) (*protos.VerifyCredentialResponse, error) {
	request := &protos.VerifyCredentialRequest{
		Credential: cr,
	}
	switch providerName := strings.ToLower(providerName); providerName {
	case "cohere":
		return client.cohereClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "anthropic":
		return client.anthropicClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "replicate":
		return client.replicateClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "gemini":
		return client.geminiClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "vertexai":
		return client.vertexaiClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "mistral":
		return client.mistralClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "openai":
		return client.openAiClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "groq":
		// Groq uses OpenAI-compatible API, reuse OpenAI client
		return client.openAiClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "voyageai":
		return client.voyageAiClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "huggingface":
		return client.huggingfaceClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "aws-bedrock":
		return client.awsbedrockClient.VerifyCredential(client.WithAuth(c, auth), request)
	case "azure-foundry":
		return client.azureAiClient.VerifyCredential(client.WithAuth(c, auth), request)
	default:
		return nil, errors.New("illegal provider for chat request")
	}
}

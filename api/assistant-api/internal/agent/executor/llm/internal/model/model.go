// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_model

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	internal_agent_executor "github.com/rapidaai/api/assistant-api/internal/agent/executor"
	internal_agent_tool "github.com/rapidaai/api/assistant-api/internal/agent/executor/tool"
	internal_adapter_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	integration_client_builders "github.com/rapidaai/pkg/clients/integration/builders"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"golang.org/x/sync/errgroup"
)

type modelAssistantExecutor struct {
	logger             commons.Logger
	toolExecutor       internal_agent_executor.ToolExecutor
	providerCredential *protos.VaultCredential
	inputBuilder       integration_client_builders.InputChatBuilder
	history            []*protos.Message
}

func NewModelAssistantExecutor(logger commons.Logger) internal_agent_executor.AssistantExecutor {
	return &modelAssistantExecutor{
		logger:       logger,
		inputBuilder: integration_client_builders.NewChatInputBuilder(logger),
		toolExecutor: internal_agent_tool.NewToolExecutor(logger),
		history:      make([]*protos.Message, 0),
	}

}

func (executor *modelAssistantExecutor) Name() string {
	return "model"
}

func (executor *modelAssistantExecutor) Initialize(ctx context.Context, communication internal_type.Communication, cfg *protos.AssistantConversationConfiguration) error {
	start := time.Now()
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantAgentConnectStage, internal_adapter_telemetry.KV{K: "executor", V: internal_adapter_telemetry.StringValue(executor.Name())})
	defer span.EndSpan(ctx, utils.AssistantAgentConnectStage)

	g, gCtx := errgroup.WithContext(ctx)

	var providerCredential *protos.VaultCredential
	var conversationLogs []*protos.Message

	// Goroutine to fetch provider credentials
	g.Go(func() error {
		credentialID, err := communication.Assistant().AssistantProviderModel.GetOptions().GetUint64("rapida.credential_id")
		if err != nil {
			executor.logger.Errorf("Error while getting provider model credential ID: %v", err)
			return fmt.Errorf("failed to get credential ID: %w", err)
		}
		span.AddAttributes(gCtx, internal_adapter_telemetry.KV{K: "vault_id", V: internal_adapter_telemetry.IntValue(credentialID)})

		cred, err := communication.VaultCaller().GetCredential(gCtx, communication.Auth(), credentialID)
		if err != nil {
			executor.logger.Errorf("Error while getting provider model credentials: %v", err)
			return fmt.Errorf("failed to get provider credential: %w", err)
		}
		providerCredential = cred
		return nil
	})

	// Goroutine to fetch conversation logs
	g.Go(func() error {
		conversationLogs = communication.GetConversationLogs()
		return nil
	})

	// Goroutine to initialize tool executor
	g.Go(func() error {
		if err := executor.toolExecutor.Initialize(gCtx, communication); err != nil {
			executor.logger.Errorf("Error initializing tool executor: %v", err)
			return fmt.Errorf("failed to initialize tool executor: %w", err)
		}
		return nil
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		executor.logger.Errorf("Error during initialization: %v", err)
		return err
	}

	// Assign after goroutines complete to avoid race conditions
	executor.providerCredential = providerCredential
	executor.history = append(executor.history, conversationLogs...)
	span.AddAttributes(ctx, internal_adapter_telemetry.KV{K: "history_length", V: internal_adapter_telemetry.IntValue(len(executor.history))})

	executor.logger.Benchmark("DefaultAssistantExecutor.Init", time.Since(start))
	return nil
}

func (executor *modelAssistantExecutor) chat(
	ctx context.Context,
	communication internal_type.Communication,
	packet internal_type.LLMMessagePacket,
	histories ...*protos.Message,
) error {
	request := executor.buildChatRequest(communication, packet, histories...)
	res, err := communication.IntegrationCaller().StreamChat(
		ctx,
		communication.Auth(),
		communication.Assistant().AssistantProviderModel.ModelProviderName,
		request,
	)
	if err != nil {
		executor.logger.Errorf("error while streaming chat request: %v", err)
		return fmt.Errorf("failed to stream chat: %w", err)
	}

	return executor.processStream(ctx, communication, packet, res, histories)
}

// buildChatRequest constructs the chat request with all necessary parameters
func (executor *modelAssistantExecutor) buildChatRequest(communication internal_type.Communication, packet internal_type.LLMMessagePacket, histories ...*protos.Message) *protos.ChatRequest {
	assistant := communication.Assistant()
	template := assistant.AssistantProviderModel.Template.GetTextChatCompleteTemplate()
	messages := executor.inputBuilder.Message(
		template.Prompt,
		utils.MergeMaps(executor.inputBuilder.PromptArguments(template.Variables), communication.GetArgs()),
	)
	messages = append(messages, histories...)
	messages = append(messages, packet.Message.ToProto())

	return executor.inputBuilder.Chat(
		&protos.Credential{
			Id:    executor.providerCredential.GetId(),
			Value: executor.providerCredential.GetValue(),
		},
		executor.inputBuilder.Options(
			utils.MergeMaps(assistant.AssistantProviderModel.GetOptions(), communication.GetOptions()),
			nil,
		),
		executor.toolExecutor.GetFunctionDefinitions(),
		map[string]string{
			"assistant_id":                fmt.Sprintf("%d", assistant.Id),
			"assistant_provider_model_id": fmt.Sprintf("%d", assistant.AssistantProviderModel.Id),
		},
		messages...,
	)
}

// processStream handles the streaming response from the LLM
func (executor *modelAssistantExecutor) processStream(
	ctx context.Context,
	communication internal_type.Communication,
	packet internal_type.LLMMessagePacket,
	res protos.OpenAiService_StreamChatClient,
	histories []*protos.Message,
) error {
	var (
		output  *protos.Message
		metrics []*protos.Metric
	)

	for {
		msg, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return executor.handleStreamComplete(ctx, communication, packet, output, metrics, histories)
			}
			return fmt.Errorf("failed to receive stream message: %w", err)
		}

		metrics = msg.GetMetrics()
		output = msg.GetData()

		if metrics != nil {
			// Metrics available means end of generation
			communication.OnPacket(ctx, internal_type.MetricPacket{
				ContextID: packet.ContextID,
				Metrics:   types.ToMetrics(metrics),
			})
			continue
		}

		if output != nil && len(output.GetContents()) > 0 {
			communication.OnPacket(ctx, internal_type.LLMStreamPacket{
				ContextID: packet.ContextID,
				Text:      types.ToMessage(output).String(),
			})
		}
	}
}

// handleStreamComplete processes the final output when stream ends
func (executor *modelAssistantExecutor) handleStreamComplete(
	ctx context.Context,
	communication internal_type.Communication,
	packet internal_type.LLMMessagePacket,
	output *protos.Message,
	metrics []*protos.Metric,
	histories []*protos.Message,
) error {
	outputMessage := types.ToMessage(output)
	outPacket := internal_type.LLMMessagePacket{
		ContextID: packet.ContextID,
		Message:   outputMessage,
	}
	metricPacket := internal_type.MetricPacket{
		ContextID: packet.ContextID,
		Metrics:   types.ToMetrics(metrics),
	}

	executor.recordLLMInteraction(communication, packet, outPacket, metricPacket)
	communication.OnPacket(ctx, outPacket)

	// Handle tool calls if present
	if output != nil && len(output.GetToolCalls()) > 0 {
		return executor.executeToolCalls(ctx, communication, packet, output, histories)
	}

	return nil
}

// executeToolCalls handles tool execution and recursive chat
func (executor *modelAssistantExecutor) executeToolCalls(ctx context.Context, communication internal_type.Communication, packet internal_type.LLMMessagePacket, output *protos.Message, histories []*protos.Message,
) error {
	toolExecution, toolContents := executor.toolExecutor.ExecuteAll(ctx, packet, output.GetToolCalls(), communication)
	updatedHistories := append(histories, packet.Message.ToProto(), output)
	communication.OnPacket(ctx, toolExecution...)
	err := executor.chat(ctx, communication, internal_type.LLMMessagePacket{ContextID: packet.ContextID, Message: &types.Message{Contents: toolContents, Role: "tool"}}, updatedHistories...)
	return err
}

// recordLLMInteraction appends messages to history and persists to storage
func (executor *modelAssistantExecutor) recordLLMInteraction(communication internal_type.Communication, in, out internal_type.LLMMessagePacket, metrics internal_type.MetricPacket,
) {
	if in.Message != nil {
		executor.history = append(executor.history, in.Message.ToProto())
	}
	if out.Message != nil {
		executor.history = append(executor.history, out.Message.ToProto())
	}

	// Persist to storage asynchronously
	utils.Go(context.Background(), func() {
		communication.CreateConversationMessageLog(in.ContextID, in.Message, out.Message, metrics.Metrics)
	})
}

// Execute processes incoming packets when user triggers a message
func (executor *modelAssistantExecutor) Execute(ctx context.Context, communication internal_type.Communication, pctk internal_type.Packet) error {
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantAgentTextGenerationStage, internal_adapter_telemetry.MessageKV(pctk.ContextId()))
	defer span.EndSpan(ctx, utils.AssistantAgentTextGenerationStage)
	switch plt := pctk.(type) {
	case internal_type.UserTextPacket:
		return executor.handleUserTextPacket(ctx, communication, plt)
	case internal_type.StaticPacket:
		return executor.handleStaticPacket(plt)
	default:
		return fmt.Errorf("unsupported packet type: %T", pctk)
	}
}

// handleUserTextPacket processes user text input
func (executor *modelAssistantExecutor) handleUserTextPacket(ctx context.Context, communication internal_type.Communication, packet internal_type.UserTextPacket,
) error {
	message := types.NewMessage("user", &types.Content{ContentType: commons.TEXT_CONTENT.String(), ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(), Content: []byte(packet.Text)})
	return executor.chat(ctx, communication, internal_type.LLMMessagePacket{ContextID: packet.ContextId(), Message: message}, executor.history...)
}

// handleStaticPacket appends static assistant response to history
func (executor *modelAssistantExecutor) handleStaticPacket(packet internal_type.StaticPacket) error {
	executor.history = append(executor.history, &protos.Message{
		Role: "assistant",
		Contents: []*protos.Content{
			{
				ContentType:   commons.TEXT_CONTENT.String(),
				ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
				Content:       []byte(packet.Text),
			},
		},
	})
	return nil
}

func (executor *modelAssistantExecutor) Close(ctx context.Context, communication internal_type.Communication) error {
	executor.history = make([]*protos.Message, 0)
	return nil
}

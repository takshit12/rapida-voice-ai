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
	internal_executors "github.com/rapidaai/api/assistant-api/internal/agent/executor"
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
	toolExecutor       internal_executors.ToolExecutor
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

func (executor *modelAssistantExecutor) Initialize(ctx context.Context, communication internal_type.Communication) error {
	start := time.Now()
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantAgentConnectStage, internal_adapter_telemetry.KV{K: "executor", V: internal_adapter_telemetry.StringValue(executor.Name())})
	defer span.EndSpan(ctx, utils.AssistantAgentConnectStage)
	g, ctx := errgroup.WithContext(ctx)

	var providerCredential *protos.VaultCredential
	g.Go(func() error {
		credentialId, err := communication.Assistant().AssistantProviderModel.GetOptions().GetUint64("rapida.credential_id")
		if err != nil {
			executor.logger.Errorf("Error while getting provider model credential ID: %v", err)
			return fmt.Errorf("failed to get credential ID: %w", err)
		}
		span.AddAttributes(ctx, internal_adapter_telemetry.KV{K: "vault_id", V: internal_adapter_telemetry.IntValue(credentialId)})
		providerCredential, err = communication.VaultCaller().GetCredential(ctx, communication.Auth(), credentialId)
		if err != nil {
			executor.logger.Errorf("Error while getting provider model credentials: %v", err)
			return fmt.Errorf("failed to get provider credential: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		executor.history = append(executor.history, communication.GetConversationLogs()...)
		span.AddAttributes(ctx, internal_adapter_telemetry.KV{K: "history_length", V: internal_adapter_telemetry.IntValue(len(executor.history))})
		return nil

	})
	// Goroutine to initialize tool executor
	g.Go(func() error {
		if err := executor.toolExecutor.Initialize(ctx, communication); err != nil {
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
	executor.providerCredential = providerCredential
	executor.logger.Benchmark("DefaultAssistantExecutor.Init", time.Since(start))
	return nil
}

func (executor *modelAssistantExecutor) chat(
	ctx context.Context,

	// for communication
	communication internal_type.Communication,

	// llm packet
	packet internal_type.LLMPacket,

	// histories or older conversation
	histories ...*protos.Message,
) error {
	var (
		output  *protos.Message
		metrics []*protos.Metric
	)
	request := executor.inputBuilder.Chat(
		&protos.Credential{
			Id:    executor.providerCredential.GetId(),
			Value: executor.providerCredential.GetValue(),
		},
		executor.inputBuilder.Options(utils.MergeMaps(communication.Assistant().AssistantProviderModel.GetOptions(), communication.GetOptions()), nil),
		executor.toolExecutor.GetFunctionDefinitions(),
		map[string]string{
			"assistant_id":                fmt.Sprintf("%d", communication.Assistant().Id),
			"assistant_provider_model_id": fmt.Sprintf("%d", communication.Assistant().AssistantProviderModel.Id),
		},
		append(append(
			executor.inputBuilder.Message(communication.Assistant().AssistantProviderModel.Template.GetTextChatCompleteTemplate().Prompt, utils.MergeMaps(executor.inputBuilder.PromptArguments(communication.Assistant().AssistantProviderModel.Template.GetTextChatCompleteTemplate().Variables), communication.GetArgs())), histories...), packet.Message.ToProto())...,
	)

	res, err := communication.IntegrationCaller().StreamChat(ctx, communication.Auth(), communication.Assistant().AssistantProviderModel.ModelProviderName, request)
	if err != nil {
		executor.logger.Errorf("error while streaming chat request: %v", err)
		return err
	}
	for {
		msg, err := res.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				executor.llm(communication, packet, internal_type.LLMPacket{ContextID: packet.ContextId(), Message: types.ToMessage(output)}, internal_type.MetricPacket{ContextID: packet.ContextID, Metrics: types.ToMetrics(metrics)})
				communication.OnPacket(ctx, internal_type.LLMPacket{ContextID: packet.ContextId(), Message: types.ToMessage(output)})
				if len(output.GetToolCalls()) > 0 {
					// append history of tool call
					toolExecution, toolContents := executor.toolExecutor.ExecuteAll(ctx, packet, output.GetToolCalls(), communication)
					communication.OnPacket(ctx, toolExecution...)
					return executor.chat(ctx, communication,
						internal_type.LLMPacket{ContextID: packet.ContextId(), Message: &types.Message{Contents: toolContents, Role: "tool"}},
						append(histories, packet.Message.ToProto(), output)...)
				}
				return nil
			}
			return err
		}

		metrics = msg.GetMetrics()
		output = msg.GetData()
		if metrics != nil {
			// metrics available means end of generation
			communication.OnPacket(ctx, internal_type.MetricPacket{ContextID: packet.ContextID, Metrics: types.ToMetrics(metrics)})
			continue
		}
		if output != nil && len(output.GetContents()) > 0 {
			communication.OnPacket(ctx, internal_type.LLMStreamPacket{ContextID: packet.ContextId(), Message: types.ToMessage(msg.GetData())})
		}

	}
}

func (executor *modelAssistantExecutor) llm(communication internal_type.Communication, in, out internal_type.LLMPacket, metrics internal_type.MetricPacket) error {
	if in.Message != nil {
		executor.history = append(executor.history, in.Message.ToProto())
	}
	if out.Message != nil {
		executor.history = append(executor.history, out.Message.ToProto())
	}
	// persist it to storage
	utils.Go(context.Background(), func() {
		communication.CreateConversationMessageLog(in.ContextID, in.Message, out.Message, metrics.Metrics)
	})
	return nil
}

// when user tigger a message
func (executor *modelAssistantExecutor) Execute(ctx context.Context, communication internal_type.Communication, pctk internal_type.Packet) error {
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantAgentTextGenerationStage, internal_adapter_telemetry.MessageKV(pctk.ContextId()))
	defer span.EndSpan(ctx, utils.AssistantAgentTextGenerationStage)
	switch plt := pctk.(type) {
	case internal_type.UserTextPacket:
		return executor.chat(ctx, communication, internal_type.LLMPacket{ContextID: pctk.ContextId(), Message: types.NewMessage("user", &types.Content{
			ContentType:   commons.TEXT_CONTENT.String(),
			ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
			Content:       []byte(plt.Text),
		})}, executor.history...)
	case internal_type.StaticPacket:
		executor.history = append(executor.history, &protos.Message{
			Role: "assistant",
			Contents: []*protos.Content{
				{
					ContentType:   commons.TEXT_CONTENT.String(),
					ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
					Content:       []byte(plt.Text),
				},
			},
		})
		return nil
	default:
		return fmt.Errorf("unsupported packet type: %T", pctk)
	}

}

func (executor *modelAssistantExecutor) Close(ctx context.Context, communication internal_type.Communication) error {
	executor.history = make([]*protos.Message, 0)
	return nil
}

// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_agentkit

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	internal_agent_executor "github.com/rapidaai/api/assistant-api/internal/agent/executor"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_adapter_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/types"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type agentkitExecutor struct {
	logger   commons.Logger
	agentKit protos.AgentTalkClient
	talker   grpc.BidiStreamingClient[protos.AgentTalkRequest, protos.AgentTalkResponse]
	conn     *grpc.ClientConn
	mu       sync.RWMutex
}

// NewAgentKitAssistantExecutor creates a new AgentKit-based assistant executor
// that communicates with an external gRPC service for LLM operations.
func NewAgentKitAssistantExecutor(logger commons.Logger) internal_agent_executor.AssistantExecutor {
	return &agentkitExecutor{
		logger: logger,
	}
}

// Name returns the executor name identifier.
func (executor *agentkitExecutor) Name() string {
	return "agentkit"
}

// Initialize establishes the gRPC connection to the AgentKit service and starts
// the bidirectional stream for communication.
func (executor *agentkitExecutor) Initialize(ctx context.Context, communication internal_type.Communication, cfg *protos.AssistantConversationConfiguration) error {
	start := time.Now()
	ctx, span, _ := communication.Tracer().StartSpan(ctx, utils.AssistantAgentConnectStage, internal_adapter_telemetry.KV{K: "executor", V: internal_adapter_telemetry.StringValue(executor.Name())})
	defer span.EndSpan(ctx, utils.AssistantAgentConnectStage)

	providerDefinition := communication.Assistant().AssistantProviderAgentkit
	if providerDefinition == nil {
		return fmt.Errorf("agentkit provider definition is nil")
	}

	g, gCtx := errgroup.WithContext(ctx)

	// Goroutine to establish gRPC connection and start streaming
	g.Go(func() error {
		return executor.establishConnection(gCtx, providerDefinition, communication)
	})

	if err := g.Wait(); err != nil {
		executor.logger.Errorf("Error during initialization of agentkit: %v", err)
		return err
	}

	// Start the response listener in background
	utils.Go(ctx, func() {
		if err := executor.responseListener(ctx, communication); err != nil {
			executor.logger.Errorf("Error in AgentKit response listener: %v", err)
		}
	})

	// Send initial configuration to AgentKit
	if err := executor.sendConfiguration(communication, cfg); err != nil {
		return fmt.Errorf("failed to send configuration: %w", err)
	}
	executor.logger.Benchmark("AgentKitExecutor.Initialize", time.Since(start))
	return nil
}

// establishConnection creates the gRPC connection with proper TLS/credentials configuration.
func (executor *agentkitExecutor) establishConnection(
	ctx context.Context,
	provider *internal_assistant_entity.AssistantProviderAgentkit,
	communication internal_type.Communication,
) error {
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(math.MaxInt64),
			grpc.MaxCallSendMsgSize(math.MaxInt64),
		),
	}

	// Configure TLS if certificate is provided
	if provider.Certificate != "" {
		tlsCreds, err := executor.buildTLSCredentials(provider.Certificate)
		if err != nil {
			return fmt.Errorf("failed to build TLS credentials: %w", err)
		}
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(tlsCreds))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(provider.Url, grpcOpts...)
	if err != nil {
		return fmt.Errorf("failed to connect to agentkit: %w", err)
	}
	executor.conn = conn
	executor.agentKit = protos.NewAgentTalkClient(conn)

	// Build metadata from provider configuration
	md := metadata.New(provider.Metadata)
	streamCtx := metadata.NewOutgoingContext(ctx, md)

	talker, err := executor.agentKit.Talk(streamCtx)
	if err != nil {
		return fmt.Errorf("failed to start agentkit stream: %w", err)
	}
	executor.talker = talker
	return nil
}

// buildTLSCredentials creates TLS credentials from a PEM certificate.
func (executor *agentkitExecutor) buildTLSCredentials(certPEM string) (credentials.TransportCredentials, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(certPEM)) {
		return nil, fmt.Errorf("failed to parse certificate")
	}
	return credentials.NewTLS(&tls.Config{
		RootCAs: certPool,
	}), nil
}

// sendConfiguration sends the initial configuration to the AgentKit service.
func (executor *agentkitExecutor) sendConfiguration(communication internal_type.Communication, cfg *protos.AssistantConversationConfiguration) error {
	if executor.talker == nil {
		return fmt.Errorf("talker stream is not initialized")
	}

	return executor.talker.Send(&protos.AgentTalkRequest{
		Request: &protos.AgentTalkRequest_Configuration{
			Configuration: &protos.AssistantConversationConfiguration{
				AssistantConversationId: communication.Conversation().Id,
				Assistant: &protos.AssistantDefinition{
					AssistantId: communication.Assistant().Id,
					Version:     utils.GetVersionString(communication.Assistant().AssistantProviderId),
				},
				Args:         cfg.GetArgs(),
				Metadata:     cfg.GetMetadata(),
				Options:      cfg.GetOptions(),
				InputConfig:  cfg.GetInputConfig(),
				OutputConfig: cfg.GetOutputConfig(),
			},
		},
	})
}

// responseListener listens for responses from the AgentKit service and processes them.
func (executor *agentkitExecutor) responseListener(ctx context.Context, communication internal_type.Communication) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if executor.talker == nil {
			return fmt.Errorf("talker stream is nil")
		}

		resp, err := executor.talker.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) || status.Code(err) == codes.Canceled {
				executor.logger.Debugf("AgentKit stream closed")
				return nil
			}
			return fmt.Errorf("stream recv error: %w", err)
		}

		if err := executor.processResponse(ctx, resp, communication); err != nil {
			executor.logger.Errorf("Error processing AgentKit response: %v", err)
		}
	}
}

// processResponse handles individual responses from the AgentKit service.
func (executor *agentkitExecutor) processResponse(
	ctx context.Context,
	resp *protos.AgentTalkResponse,
	communication internal_type.Communication,
) error {
	if resp.GetError() != nil {
		executor.logger.Errorf("AgentKit error response: code=%d, message=%s", resp.GetCode())
		return nil
	}

	if !resp.GetSuccess() {
		return nil
	}

	switch data := resp.GetData().(type) {
	case *protos.AgentTalkResponse_Interruption:
		communication.OnPacket(ctx, internal_type.InterruptionPacket{
			ContextID: fmt.Sprintf("%d", communication.Conversation().Id),
			Source:    internal_type.InterruptionSourceWord,
		})

	case *protos.AgentTalkResponse_Assistant:
		if err := executor.processAssistantMessage(ctx, data.Assistant, communication); err != nil {
			return err
		}
	}
	return nil
}

// processAssistantMessage handles assistant messages from AgentKit.
func (executor *agentkitExecutor) processAssistantMessage(
	ctx context.Context,
	assistant *protos.AssistantConversationAssistantMessage,
	communication internal_type.Communication,
) error {
	if assistant == nil {
		return nil
	}

	contextID := assistant.GetId()
	if contextID == "" {
		contextID = fmt.Sprintf("%d", communication.Conversation().Id)
	}

	switch msg := assistant.GetMessage().(type) {
	case *protos.AssistantConversationAssistantMessage_Text:
		// Send streaming text packet
		communication.OnPacket(ctx, internal_type.LLMStreamPacket{
			ContextID: contextID,
			Text:      msg.Text.GetContent(),
		})

		// If completed, send the full message packet
		if assistant.GetCompleted() {
			message := types.NewMessage("assistant", &types.Content{
				ContentType:   commons.TEXT_CONTENT.String(),
				ContentFormat: commons.TEXT_CONTENT_FORMAT_RAW.String(),
				Content:       []byte(msg.Text.GetContent()),
			})
			communication.OnPacket(ctx, internal_type.LLMMessagePacket{
				ContextID: contextID,
				Message:   message,
			})
		}

	case *protos.AssistantConversationAssistantMessage_Audio:
		// Handle audio messages if needed
		executor.logger.Debugf("Received audio message from AgentKit (not implemented)")
	}

	return nil
}

// Execute processes incoming packets and sends them to the AgentKit service.
func (executor *agentkitExecutor) Execute(
	ctx context.Context,
	communication internal_type.Communication,
	packet internal_type.Packet,
) error {
	ctx, span, _ := communication.Tracer().StartSpan(
		ctx,
		utils.AssistantAgentTextGenerationStage,
		internal_adapter_telemetry.MessageKV(packet.ContextId()),
	)
	defer span.EndSpan(ctx, utils.AssistantAgentTextGenerationStage)

	switch p := packet.(type) {
	case internal_type.UserTextPacket:
		return executor.handleUserTextPacket(ctx, p, communication)
	case internal_type.StaticPacket:
		return executor.handleStaticPacket(p)
	default:
		return fmt.Errorf("unsupported packet type: %T", packet)
	}
}

// handleUserTextPacket sends user text messages to the AgentKit service.
func (executor *agentkitExecutor) handleUserTextPacket(
	ctx context.Context,
	packet internal_type.UserTextPacket,
	communication internal_type.Communication,
) error {
	if executor.talker == nil {
		return fmt.Errorf("talker stream is not initialized")
	}

	// Send message to AgentKit
	return executor.talker.Send(&protos.AgentTalkRequest{
		Request: &protos.AgentTalkRequest_Message{
			Message: &protos.AssistantConversationUserMessage{
				Message: &protos.AssistantConversationUserMessage_Text{
					Text: &protos.AssistantConversationMessageTextContent{
						Content: packet.Text,
					},
				},
				Id:        packet.ContextId(),
				Completed: true,
				Time:      timestamppb.Now(),
			},
		},
	})
}

// handleStaticPacket appends static assistant responses to history.
func (executor *agentkitExecutor) handleStaticPacket(packet internal_type.StaticPacket) error {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	return nil
}

// Close terminates the AgentKit connection and cleans up resources.
func (executor *agentkitExecutor) Close(ctx context.Context, communication internal_type.Communication) error {
	executor.logger.Debugf("Closing AgentKit executor")

	if executor.talker != nil {
		if err := executor.talker.CloseSend(); err != nil {
			executor.logger.Errorf("Error closing talker stream: %v", err)
		}
		executor.talker = nil
	}

	if executor.conn != nil {
		if err := executor.conn.Close(); err != nil {
			executor.logger.Errorf("Error closing gRPC connection: %v", err)
		}
		executor.conn = nil
	}
	return nil
}

// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

// Package internal_adapter_generic provides the generic adapter implementation
// for managing voice assistant sessions. It handles the complete lifecycle of
// assistant conversations including connection, disconnection, audio streaming,
// and state management.
package internal_adapter_generic

import (
	"context"
	"fmt"
	"sync"
	"time"

	internal_adapter_request_customizers "github.com/rapidaai/api/assistant-api/internal/adapters/customizers"
	internal_audio_recorder "github.com/rapidaai/api/assistant-api/internal/audio/recorder"
	internal_assistant_entity "github.com/rapidaai/api/assistant-api/internal/entity/assistants"
	internal_conversation_gorm "github.com/rapidaai/api/assistant-api/internal/entity/conversations"
	internal_telemetry "github.com/rapidaai/api/assistant-api/internal/telemetry"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/types"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// clientInfoMetadataKey is the metadata key used to store client information.
	clientInfoMetadataKey = "talk.client_information"

	// versionPrefix is the prefix used for assistant version identifiers.
	versionPrefix = "vrsn_%d"
)

// =============================================================================
// Session Lifecycle Management
// =============================================================================

// Disconnect gracefully terminates an active assistant conversation session.
//
// This method orchestrates the complete disconnection lifecycle by:
//   - Closing all active listeners (speech-to-text transformers)
//   - Closing all active speakers (text-to-speech transformers)
//   - Flushing final conversation metrics (duration, status)
//   - Persisting audio recordings to storage
//   - Exporting telemetry data for analytics
//   - Cleaning up the assistant executor
//   - Stopping any active idle timeout timers
//
// The method executes resource cleanup operations concurrently for optimal
// performance while ensuring all operations complete before returning.
//
// Thread Safety: This method is safe to call from any goroutine, but should
// only be called once per session to avoid duplicate cleanup operations.
func (r *GenericRequestor) Disconnect() {
	ctx, span, _ := r.Tracer().StartSpan(r.Context(), utils.AssistantDisconnectStage)
	startTime := time.Now()

	// Phase 1: Close all session resources concurrently
	r.closeSessionResources(ctx)

	// Phase 2: Trigger end-of-conversation hooks
	r.OnEndConversation()

	// Phase 3: Persist audio recording asynchronously
	r.persistRecording(ctx)

	// Phase 4: Complete the tracing span
	span.EndSpan(ctx, utils.AssistantDisconnectStage)

	// Phase 5: Export telemetry and cleanup
	r.exportTelemetry(ctx)
	r.closeExecutor(ctx)
	r.stopTimers()

	r.logger.Benchmark("session.Disconnect", time.Since(startTime))
}

// Connect establishes a new assistant session or resumes an existing one.
//
// This method serves as the primary entry point for initiating assistant
// conversations. Based on the provided configuration, it either creates
// a new session or resumes an existing conversation.
//
// Parameters:
//   - ctx: Context for cancellation and deadline propagation
//   - auth: Authentication principal containing user/organization credentials
//   - identifier: Unique session identifier (e.g., phone number, client ID)
//   - config: Conversation configuration including assistant details and audio settings
//
// Returns:
//   - error: nil on success, or an error describing the failure reason
//
// The method performs the following steps:
//  1. Validates and creates the request customizer
//  2. Sets authentication context
//  3. Retrieves the assistant configuration
//  4. Routes to either new session creation or session resumption
//
// Example:
//
//	err := requestor.Connect(ctx, auth, "user-123", &protos.AssistantConversationConfiguration{
//	    Assistant: &protos.AssistantDefinition{AssistantId: 1},
//	})
func (r *GenericRequestor) Connect(
	ctx context.Context,
	auth types.SimplePrinciple,
	identifier string,
	config *protos.AssistantConversationConfiguration,
) error {
	ctx, span, _ := r.Tracer().StartSpan(ctx, utils.AssistantConnectStage)
	defer span.EndSpan(ctx, utils.AssistantConnectStage)

	// Create request customizer from configuration
	customizer, err := internal_adapter_request_customizers.NewRequestBaseCustomizer(config)
	if err != nil {
		r.logger.Errorf("failed to initialize request customizer: %+v", err)
		return err
	}

	// Set authentication context
	r.SetAuth(auth)

	// Retrieve assistant configuration
	assistant, err := r.GetAssistant(auth, config.Assistant.AssistantId, config.Assistant.Version)
	if err != nil {
		r.logger.Errorf("failed to retrieve assistant configuration: %+v", err)
		return err
	}

	// Route to appropriate session handler based on conversation ID presence
	if conversationID := config.GetAssistantConversationId(); conversationID > 0 {
		span.AddAttributes(ctx,
			internal_telemetry.KV{K: "conversation_initiation", V: internal_telemetry.StringValue("resume")},
			internal_telemetry.KV{K: "conversation_id", V: internal_telemetry.IntValue(conversationID)},
		)
		return r.resumeSession(ctx, config, assistant, identifier, customizer)
	}

	span.AddAttributes(ctx,
		internal_telemetry.KV{K: "conversation_initiation", V: internal_telemetry.StringValue("new")},
	)
	return r.createSession(ctx, config, assistant, identifier, customizer)
}

// OnCreateSession initializes a new assistant conversation session.
//
// This method sets up all necessary components for a new voice conversation:
//   - Creates a new conversation record in the database
//   - Configures audio input/output modes based on stream settings
//   - Initializes the LLM executor for message processing
//   - Establishes speaker and listener connections
//   - Starts audio recording if enabled
//   - Sends initial configuration to the client
//
// Parameters:
//   - ctx: Context for cancellation and deadline propagation
//   - inputConfig: Stream configuration for audio input (microphone)
//   - outputConfig: Stream configuration for audio output (speaker)
//   - assistant: The assistant entity containing configuration details
//   - identifier: Unique session identifier
//   - customizer: Request customization options (args, metadata, options)
//
// Returns:
//   - error: nil on success, or an error if session creation fails
//
// Concurrency: This method spawns multiple goroutines for parallel
// initialization. Critical path operations run in the error group,
// while non-critical operations run as background tasks.
func (r *GenericRequestor) OnCreateSession(
	ctx context.Context,
	inputConfig, outputConfig *protos.StreamConfig,
	assistant *internal_assistant_entity.Assistant,
	identifier string,
	customizer internal_type.Customization,
) error {
	ctx, span, _ := r.Tracer().StartSpan(ctx, utils.AssistantCreateConversationStage)
	defer span.EndSpan(ctx, utils.AssistantCreateConversationStage)

	// Create new conversation record
	conversation, err := r.BeginConversation(
		r.Auth(),
		assistant,
		type_enums.DIRECTION_INBOUND,
		identifier,
		customizer.GetArgs(),
		customizer.GetMetadata(),
		customizer.GetOptions(),
	)
	if err != nil {
		r.logger.Errorf("failed to begin conversation: %+v", err)
		return err
	}

	// Configure audio modes based on stream settings
	audioInputConfig, audioOutputConfig := r.configureAudioModes(inputConfig, outputConfig)

	// Initialize critical components concurrently
	errGroup, _ := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return r.initializeExecutor(ctx)
	})

	errGroup.Go(func() error {
		r.notifyConfiguration(ctx, conversation, assistant)
		return nil
	})

	errGroup.Go(func() error {
		r.connectSpeakerAndInitializeBehavior(ctx, audioOutputConfig)
		return nil
	})

	// Start non-critical background tasks
	r.startBackgroundTasks(ctx, audioInputConfig, audioOutputConfig, true)

	return errGroup.Wait()
}

// OnResumeSession resumes an existing assistant conversation session.
//
// This method restores a previously active conversation, allowing users
// to continue where they left off. It re-establishes all necessary
// connections while preserving conversation history and state.
//
// Parameters:
//   - ctx: Context for cancellation and deadline propagation
//   - inputConfig: Stream configuration for audio input (microphone)
//   - outputConfig: Stream configuration for audio output (speaker)
//   - assistant: The assistant entity containing configuration details
//   - identifier: Unique session identifier
//   - conversationID: The ID of the conversation to resume
//   - customizer: Request customization options
//
// Returns:
//   - error: nil on success, or an error if session resumption fails
//
// Note: The conversation must exist and be in a resumable state.
// Attempting to resume a completed or cancelled conversation will fail.
func (r *GenericRequestor) OnResumeSession(
	ctx context.Context,
	inputConfig, outputConfig *protos.StreamConfig,
	assistant *internal_assistant_entity.Assistant,
	identifier string,
	conversationID uint64,
	customizer internal_type.Customization,
) error {
	ctx, span, _ := r.Tracer().StartSpan(r.Context(), utils.AssistantResumeConverstaionStage)
	defer span.EndSpan(ctx, utils.AssistantResumeConverstaionStage)

	// Resume existing conversation
	conversation, err := r.ResumeConversation(r.Auth(), assistant, conversationID, identifier)
	if err != nil {
		r.logger.Errorf("failed to resume conversation: %+v", err)
		return err
	}

	// Configure audio modes based on stream settings
	audioInputConfig, audioOutputConfig := r.configureAudioModes(inputConfig, outputConfig)

	// Initialize critical components concurrently
	errGroup, _ := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		return r.initializeExecutor(ctx)
	})

	errGroup.Go(func() error {
		r.notifyConfiguration(ctx, conversation, assistant)
		return nil
	})

	errGroup.Go(func() error {
		r.connectSpeakerAndInitializeBehavior(ctx, audioOutputConfig)
		return nil
	})

	// Start non-critical background tasks (not a new session)
	r.startBackgroundTasks(ctx, audioInputConfig, audioOutputConfig, false)

	// Trigger resume conversation hooks
	if err := r.OnResumeConversation(); err != nil {
		r.logger.Errorf("failed to execute resume conversation hooks: %v", err)
	}

	return errGroup.Wait()
}

// =============================================================================
// Disconnect Helpers
// =============================================================================

// closeSessionResources performs concurrent cleanup of all session resources.
//
// This method closes listeners, speakers, and flushes final metrics in parallel
// to minimize disconnection latency. It waits for all cleanup operations to
// complete before returning.
func (r *GenericRequestor) closeSessionResources(ctx context.Context) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	// Flush final conversation metrics
	utils.Go(r.Context(), func() {
		defer waitGroup.Done()
		r.flushFinalMetrics()
	})

	// Close speech-to-text listener
	utils.Go(r.Context(), func() {
		defer waitGroup.Done()
		if err := r.disconnectMicrophone(ctx); err != nil {
			r.logger.Tracef(ctx, "failed to close input transformer: %+v", err)
		}
	})

	// Close text-to-speech speaker
	utils.Go(r.Context(), func() {
		defer waitGroup.Done()
		if err := r.disconnectSpeaker(); err != nil {
			r.logger.Tracef(ctx, "failed to close output transformer: %+v", err)
		}
	})

	waitGroup.Wait()
}

// flushFinalMetrics records the final conversation metrics including
// total duration and completion status.
func (r *GenericRequestor) flushFinalMetrics() {
	conversationDuration := time.Since(r.StartedAt)

	metrics := []*types.Metric{
		{
			Name:        type_enums.TIME_TAKEN.String(),
			Value:       fmt.Sprintf("%d", int64(conversationDuration)),
			Description: "Total conversation duration in nanoseconds",
		},
		{
			Name:        type_enums.STATUS.String(),
			Value:       type_enums.RECORD_COMPLETE.String(),
			Description: "Final conversation status",
		},
	}

	r.AddMetrics(r.Auth(), metrics...)
}

// persistRecording saves the audio recording asynchronously.
//
// This method runs in a background goroutine to avoid blocking the
// disconnect flow. Any errors are logged but do not affect the
// disconnection process.
func (r *GenericRequestor) persistRecording(ctx context.Context) {
	if r.recorder != nil {
		utils.Go(r.Context(), func() {
			audioData, err := r.recorder.Persist()
			if err != nil {
				r.logger.Tracef(ctx, "failed to persist audio recording: %+v", err)
				return
			}
			if err = r.CreateConversationRecording(audioData); err != nil {
				r.logger.Tracef(ctx, "failed to create conversation recording record: %+v", err)
			}
		})
	}

}

// exportTelemetry exports conversation telemetry data for analytics and monitoring.
func (r *GenericRequestor) exportTelemetry(ctx context.Context) {
	exportOptions := &internal_telemetry.VoiceAgentExportOption{
		AssistantId:              r.assistant.Id,
		AssistantProviderModelId: r.assistant.AssistantProviderId,
		AssistantConversationId:  r.assistantConversation.Id,
	}

	if err := r.tracer.Export(ctx, r.auth, exportOptions); err != nil {
		r.logger.Errorf("failed to export telemetry data: %v", err)
	}
}

// closeExecutor shuts down the assistant executor and releases its resources.
func (r *GenericRequestor) closeExecutor(ctx context.Context) {
	if err := r.assistantExecutor.Close(ctx, r); err != nil {
		r.logger.Errorf("failed to close assistant executor: %v", err)
	}
}

// stopTimers stops all active timers (idle timeout and max session duration).
func (r *GenericRequestor) stopTimers() {
	if r.idleTimeoutTimer != nil {
		r.idleTimeoutTimer.Stop()
	}
	if r.maxSessionTimer != nil {
		r.maxSessionTimer.Stop()
	}
}

// =============================================================================
// Connect Helpers
// =============================================================================

// resumeSession delegates to OnResumeSession with extracted configuration values.
func (r *GenericRequestor) resumeSession(
	ctx context.Context,
	config *protos.AssistantConversationConfiguration,
	assistant *internal_assistant_entity.Assistant,
	identifier string,
	customizer internal_type.Customization,
) error {
	return r.OnResumeSession(
		ctx,
		config.GetInputConfig(),
		config.GetOutputConfig(),
		assistant,
		identifier,
		config.GetAssistantConversationId(),
		customizer,
	)
}

// createSession delegates to OnCreateSession with extracted configuration values.
func (r *GenericRequestor) createSession(
	ctx context.Context,
	config *protos.AssistantConversationConfiguration,
	assistant *internal_assistant_entity.Assistant,
	identifier string,
	customizer internal_type.Customization,
) error {
	return r.OnCreateSession(
		ctx,
		config.GetInputConfig(),
		config.GetOutputConfig(),
		assistant,
		identifier,
		customizer,
	)
}

// =============================================================================
// Session Setup Helpers
// =============================================================================

// configureAudioModes extracts audio configurations from stream settings
// and switches the messaging system to the appropriate input/output modes.
//
// Returns the audio configurations for both input (microphone) and output (speaker).
func (r *GenericRequestor) configureAudioModes(
	inputConfig, outputConfig *protos.StreamConfig,
) (audioInput *protos.AudioConfig, audioOutput *protos.AudioConfig) {
	audioInput = inputConfig.GetAudio()
	if audioInput != nil {
		r.messaging.SwitchInputMode(type_enums.AudioMode)
	}

	audioOutput = outputConfig.GetAudio()
	if audioOutput != nil {
		r.messaging.SwitchOutputMode(type_enums.AudioMode)
	}

	return audioInput, audioOutput
}

// initializeExecutor sets up the LLM executor for processing conversation messages.
func (r *GenericRequestor) initializeExecutor(ctx context.Context) error {
	if err := r.assistantExecutor.Initialize(ctx, r); err != nil {
		r.logger.Tracef(ctx, "failed to initialize executor: %+v", err)
		return err
	}
	return nil
}

// notifyConfiguration sends the initial conversation configuration to the client.
//
// This notification includes the conversation ID and assistant details,
// allowing the client to track the session.
func (r *GenericRequestor) notifyConfiguration(
	ctx context.Context,
	conversation *internal_conversation_gorm.AssistantConversation,
	assistant *internal_assistant_entity.Assistant,
) {
	configNotification := &protos.AssistantConversationConfiguration{
		AssistantConversationId: conversation.Id,
		Assistant: &protos.AssistantDefinition{
			AssistantId: assistant.Id,
			Version:     fmt.Sprintf(versionPrefix, assistant.AssistantProviderId),
		},
		Time: timestamppb.Now(),
	}

	if err := r.Notify(ctx, configNotification); err != nil {
		r.logger.Errorf("failed to send configuration notification: %v", err)
	}
}

// connectSpeakerAndInitializeBehavior establishes the speaker connection
// and initializes assistant behavior (greeting, timeouts, etc.).
func (r *GenericRequestor) connectSpeakerAndInitializeBehavior(
	ctx context.Context,
	audioOutputConfig *protos.AudioConfig,
) {
	if audioOutputConfig != nil {
		if err := r.connectSpeaker(ctx, audioOutputConfig); err != nil {
			r.logger.Tracef(ctx, "failed to connect speaker: %+v", err)
		}
	}

	if err := r.initializeBehavior(ctx); err != nil {
		r.logger.Errorf("failed to initialize assistant behavior: %+v", err)
	}
}

// startBackgroundTasks initiates non-critical background operations for session setup.
//
// These tasks run asynchronously and do not block session initialization:
//   - Audio recorder initialization
//   - Listener (speech-to-text) connection
//   - Status metric updates
//   - Client information storage
//   - Begin conversation hooks (for new sessions only)
//
// Parameters:
//   - ctx: Context for the background operations
//   - audioInputConfig: Audio configuration for the listener
//   - audioOutputConfig: Audio configuration for the recorder
//   - isNewSession: Whether this is a new session (triggers OnBeginConversation)
func (r *GenericRequestor) startBackgroundTasks(
	ctx context.Context,
	audioInputConfig, audioOutputConfig *protos.AudioConfig,
	isNewSession bool,
) {
	// Initialize audio recorder when both input and output are configured
	utils.Go(ctx, func() {
		if audioInputConfig != nil && audioOutputConfig != nil {
			rc, err := internal_audio_recorder.GetRecorder(r.logger, audioInputConfig, audioOutputConfig)
			if err != nil {
				r.logger.Tracef(ctx, "failed to initialize audio recorder: %+v", err)
				return
			}
			r.recorder = rc
		}
	})

	// Establish speech-to-text listener connection
	utils.Go(ctx, func() {
		if audioInputConfig != nil {
			if err := r.connectMicrophone(ctx, audioInputConfig); err != nil {
				r.logger.Tracef(ctx, "failed to connect listener: %+v", err)
			}
		}
	})

	// Update conversation status metric
	utils.Go(ctx, func() {
		r.AddMetrics(r.Auth(), &types.Metric{
			Name:        type_enums.STATUS.String(),
			Value:       type_enums.RECORD_IN_PROGRESS.String(),
			Description: "Conversation is currently in progress",
		})
	})

	// Store client information from gRPC context
	utils.Go(ctx, func() {
		r.storeClientInformation(ctx)
	})

	// Trigger begin conversation hooks for new sessions only
	if isNewSession {
		utils.Go(ctx, func() {
			if err := r.OnBeginConversation(); err != nil {
				r.logger.Errorf("failed to execute begin conversation hooks: %+v", err)
			}
		})
	}
}

// storeClientInformation extracts client metadata from the gRPC context
// and persists it as conversation metadata for analytics purposes.
func (r *GenericRequestor) storeClientInformation(ctx context.Context) {
	clientInfo := types.GetClientInfoFromGrpcContext(ctx)
	if clientInfo == nil {
		return
	}
	clientJSON, err := clientInfo.ToJson()
	if err != nil {
		r.logger.Tracef(ctx, "failed to serialize client information: %+v", err)
		return
	}
	r.SetMetadata(r.Auth(), map[string]interface{}{
		clientInfoMetadataKey: clientJSON,
	})
}

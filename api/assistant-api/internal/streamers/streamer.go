package internal_streamers

import (
	"context"
	"fmt"

	internal_audio "github.com/rapidaai/api/assistant-api/internal/audio"
	internal_text "github.com/rapidaai/api/assistant-api/internal/text"
	lexatic_backend "github.com/rapidaai/protos"
)

type StreamConfig struct {
	audio *internal_audio.AudioConfig `json:"audio,omitempty"`
	text  *internal_text.TextConfig   `json:"text,omitempty"`
}

func NewStreamConfig(audio *internal_audio.AudioConfig, text *internal_text.TextConfig) *StreamConfig {
	return &StreamConfig{
		audio: audio,
		text:  text,
	}
}
func (sa *StreamConfig) GetAudioConfig() (*internal_audio.AudioConfig, error) {
	if sa.audio != nil {
		return sa.audio, nil
	}
	return nil, fmt.Errorf("input audio config is not defined in stream config")
}

// StreamAttribute represents the configuration with appropriate JSON tags.
type StreamAttribute struct {
	// Input audio configuration
	inputConfig *StreamConfig `json:"input_config,omitempty"`

	// Output audio configuration
	outputConfig *StreamConfig `json:"output_config,omitempty"`
}

func NewStreamAttribute(in, out *StreamConfig) *StreamAttribute {
	return &StreamAttribute{
		inputConfig: in, outputConfig: out,
	}
}

func (sa *StreamAttribute) GetInputConfig() (*StreamConfig, error) {
	if sa.inputConfig != nil {
		return sa.inputConfig, nil
	}
	return nil, fmt.Errorf("input config is not defined in stream config")
}

func (sa *StreamAttribute) GetOutputConfig() (*StreamConfig, error) {
	if sa.outputConfig != nil {
		return sa.outputConfig, nil
	}
	return nil, fmt.Errorf("output config is not defined in stream config")
}

type Streamer interface {
	// Context returns the context associated with the stream.
	// This context is typically used to control cancellation and deadlines.
	Context() context.Context

	// Recv receives the next input value from the stream.
	// It returns the received value and any error encountered.
	// If the stream is closed, it should return (zero value, io.EOF).
	Recv() (*lexatic_backend.AssistantMessagingRequest, error)

	// Send sends an output value to the stream.
	// It returns an error if the send operation fails.
	Send(*lexatic_backend.AssistantMessagingResponse) error

	// config of streamer
	// later few more things can be added that allow more customization for source
	Config() *StreamAttribute
}

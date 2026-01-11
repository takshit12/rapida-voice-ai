// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package sarvam_internal

import (
	"encoding/json"
	"errors"
	"time"
)

type SpeechToTextTranscriptionData struct {
	RequestID  string `json:"request_id"` // Unique identifier for the request
	Transcript string `json:"transcript"` // Transcript of the provided speech in original language

	Metrics struct {
		AudioDuration     float64 `json:"audio_duration"`     // Duration of processed audio in seconds
		ProcessingLatency float64 `json:"processing_latency"` // Processing latency in seconds
	} `json:"metrics"`

	Timestamps         *interface{} `json:"timestamps,omitempty"`          // Timestamp information (if available)
	DiarizedTranscript *interface{} `json:"diarized_transcript,omitempty"` // Diarized transcript of the provided speech
	LanguageCode       *string      `json:"language_code,omitempty"`       // BCP-47 code of detected language
}

type SpeechToTextErrorData struct {
	Error string `json:"error"` // The error message
	Code  string `json:"code"`  // The error code
}

type SpeechToTextEventData struct {
	EventType  string  `json:"event_type,omitempty"`  // Optional: Type of event
	Timestamp  string  `json:"timestamp,omitempty"`   // Optional: Timestamp of the event
	SignalType string  `json:"signal_type,omitempty"` // Optional: Voice Activity Detection (VAD) signal type, e.g., "START_SPEECH", "END_SPEECH"
	OccurredAt float64 `json:"occurred_at,omitempty"` // Optional: Epoch timestamp when the event occurred
}

type SarvamSpeechToTextResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (r *SarvamSpeechToTextResponse) AsError() (*SpeechToTextErrorData, error) {
	var e SpeechToTextErrorData
	if err := json.Unmarshal(r.Data, &e); err != nil {
		return nil, err
	}
	if e.Error == "" {
		return nil, errors.New("response data is not ErrorData")
	}
	return &e, nil
}

func (r *SarvamSpeechToTextResponse) AsTranscription() (*SpeechToTextTranscriptionData, error) {
	var t SpeechToTextTranscriptionData
	if err := json.Unmarshal(r.Data, &t); err != nil {
		return nil, err
	}
	// Optional: sanity check
	if t.Transcript == "" && t.RequestID == "" {
		return nil, errors.New("not a transcription event")
	}
	return &t, nil
}

func (r *SarvamSpeechToTextResponse) AsEvent() (*SpeechToTextEventData, error) {
	var e SpeechToTextEventData
	if err := json.Unmarshal(r.Data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

type TextToSpeechErrorData struct {
	Message   string         `json:"message"`              // Required error message
	Code      *int           `json:"code,omitempty"`       // Optional error code for programmatic error handling
	Details   map[string]any `json:"details,omitempty"`    // Additional error details and context information
	RequestID *string        `json:"request_id,omitempty"` // Unique identifier for the request
}

type TextToSpeechAudioData struct {
	Audio string `json:"audio"` // Base64-encoded audio data
}

type TextToSpeechEventData struct {
	Message   string    `json:"message,omitempty"`   // Optional human-readable description
	Timestamp time.Time `json:"timestamp,omitempty"` // Optional ISO 8601 timestamp
}

type SarvamTextToSpeechResponse struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (r *SarvamTextToSpeechResponse) AsError() (*TextToSpeechErrorData, error) {
	var e TextToSpeechErrorData
	if err := json.Unmarshal(r.Data, &e); err != nil {
		return nil, err
	}
	if e.Message == "" {
		return nil, errors.New("response data is not ErrorData")
	}
	return &e, nil
}

func (r *SarvamTextToSpeechResponse) Audio() (*TextToSpeechAudioData, error) {
	var t TextToSpeechAudioData
	if err := json.Unmarshal(r.Data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// Decode error event
func (r *SarvamTextToSpeechResponse) AsEvent() (*TextToSpeechEventData, error) {
	var e TextToSpeechEventData
	if err := json.Unmarshal(r.Data, &e); err != nil {
		return nil, err
	}
	return &e, nil
}

// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package deepgram_internal

import (
	msginterfaces "github.com/deepgram/deepgram-go-sdk/v3/pkg/api/listen/v1/websocket/interfaces"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

// Implement the LiveMessageCallback interface
type deepgramSttCallback struct {
	logger   commons.Logger
	onPacket func(pkt ...internal_type.Packet) error
	options  utils.Option
}

func NewDeepgramSttCallback(
	logger commons.Logger,
	onPacket func(pkt ...internal_type.Packet) error,
	options utils.Option,
) msginterfaces.LiveMessageCallback {
	return &deepgramSttCallback{
		logger:   logger,
		onPacket: onPacket,
		options:  options,
	}

}

// Handle when the WebSocket is opened
func (d *deepgramSttCallback) Open(or *msginterfaces.OpenResponse) error {
	return nil
}

// Handle incoming transcription messages from Deepgram
func (d *deepgramSttCallback) Message(mr *msginterfaces.MessageResponse) error {
	for _, alternative := range mr.Channel.Alternatives {
		if alternative.Transcript != "" {

			if v, err := d.options.GetFloat64("listen.threshold"); err == nil {
				if float32(alternative.Confidence) < float32(v) {
					d.onPacket(
						internal_type.SpeechToTextPacket{
							Script:     alternative.Transcript,
							Confidence: alternative.Confidence,
							Language:   d.GetMostUsedLanguage(alternative.Languages),
							Interim:    true,
						},
					)
					return nil
				}
			}
			d.onPacket(
				internal_type.InterruptionPacket{Source: "word"},
				internal_type.SpeechToTextPacket{
					Script:     alternative.Transcript,
					Confidence: alternative.Confidence,
					Language:   d.GetMostUsedLanguage(alternative.Languages),
					Interim:    !mr.IsFinal,
				},
			)
			return nil
		}
	}
	return nil
}

// Handle utterance end event - this signals the end of a sentence
func (d *deepgramSttCallback) UtteranceEnd(ur *msginterfaces.UtteranceEndResponse) error {
	return nil
}

// Handle metadata (optional, can be left empty)
func (d *deepgramSttCallback) Metadata(md *msginterfaces.MetadataResponse) error {
	return nil
}

// Handle speech started event
func (d *deepgramSttCallback) SpeechStarted(ssr *msginterfaces.SpeechStartedResponse) error {
	return nil
}

// Handle when the WebSocket is closed
func (d *deepgramSttCallback) Close(cr *msginterfaces.CloseResponse) error {
	// d.logger.Debugf("Deepgram WebSocket closed")
	return nil
}

// Handle errors from Deepgram
func (d *deepgramSttCallback) Error(er *msginterfaces.ErrorResponse) error {
	d.logger.Errorf("Error %+v", er)
	return nil
}

// Handle unhandled events (optional, can be left empty)
func (d *deepgramSttCallback) UnhandledEvent(byData []byte) error {
	d.logger.Errorf("UnhandledEvent %+v", byData)
	return nil
}

func (d *deepgramSttCallback) GetMostUsedLanguage(languages []string) string {
	if len(languages) == 0 {
		return "en"
	}

	languageCount := make(map[string]int)
	for _, lang := range languages {
		languageCount[lang]++
	}

	mostUsedLang := ""
	maxCount := 0
	for lang, count := range languageCount {
		if count > maxCount {
			maxCount = count
			mostUsedLang = lang
		}
	}
	return mostUsedLang
}

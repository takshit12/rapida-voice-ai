// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_smallest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zaf/g711"

	smallest_internal "github.com/rapidaai/api/assistant-api/internal/transformer/smallest/internal"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type smallestTTS struct {
	*smallestOption
	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc

	// mutex
	mu sync.Mutex

	// current context ID from Transform calls
	currentContextID string

	logger     commons.Logger
	connection *websocket.Conn
	onPacket   func(pkt ...internal_type.Packet) error
}

func NewSmallestTextToSpeech(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential, audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option) (internal_type.TextToSpeechTransformer, error) {
	smallOpts, err := NewSmallestOption(logger, credential, audioConfig, opts)
	if err != nil {
		logger.Errorf("smallest-tts: initializing smallest failed %+v", err)
		return nil, err
	}
	ctx2, contextCancel := context.WithCancel(ctx)
	return &smallestTTS{
		ctx:            ctx2,
		ctxCancel:      contextCancel,
		onPacket:       onPacket,
		logger:         logger,
		smallestOption: smallOpts,
	}, nil
}

// Initialize implements internal_type.TextToSpeechTransformer.
func (ct *smallestTTS) Initialize() error {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+ct.GetKey())
	conn, resp, err := websocket.DefaultDialer.Dial(ct.GetTextToSpeechConnectionString(), header)
	if err != nil {
		ct.logger.Errorf("smallest-tts: error while connecting to smallest %s with response %v", err, resp)
		return err
	}

	ct.mu.Lock()
	ct.connection = conn
	defer ct.mu.Unlock()

	go ct.textToSpeechCallback(conn, ct.ctx)
	return nil
}

// Name implements internal_type.TextToSpeechTransformer.
func (*smallestTTS) Name() string {
	return "smallest-text-to-speech"
}

func (st *smallestTTS) textToSpeechCallback(conn *websocket.Conn, ctx context.Context) {
	// Smallest WS always returns PCM16; convert to mulaw if telephony requires it
	needsMulaw := st.audioConfig.GetAudioFormat() == protos.AudioConfig_MuLaw8

	for {
		select {
		case <-ctx.Done():
			st.logger.Infof("smallest-tts: context cancelled, stopping response listener")
			return
		default:
			_, audioChunk, err := conn.ReadMessage()
			if err != nil {
				if errors.Is(err, io.EOF) || websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					st.logger.Infof("smallest-tts: websocket closed gracefully")
					return
				}

				st.logger.Errorf("smallest-tts: websocket read error: %v", err)
				return
			}
			var audioData smallest_internal.SmallestTTSResponse
			if err := json.Unmarshal(audioChunk, &audioData); err != nil {
				st.logger.Errorf("smallest-tts: error parsing audio chunk: %v", err)
				continue
			}

			st.mu.Lock()
			contextID := st.currentContextID
			st.mu.Unlock()

			if contextID == "" {
				continue
			}

			st.logger.Infof("smallest-tts: received status=%s request_id=%s done=%v hasData=%v needsMulaw=%v", audioData.Status, audioData.RequestID, audioData.Done, audioData.Data != nil, needsMulaw)

			switch audioData.Status {
			case "chunk":
				if audioData.Data != nil && audioData.Data.Audio != "" {
					if rawAudioData, err := base64.StdEncoding.DecodeString(audioData.Data.Audio); err == nil {
						// Smallest returns PCM16 audio; encode to mulaw for telephony
						if needsMulaw {
							rawAudioData = g711.EncodeUlaw(rawAudioData)
						}
						st.onPacket(internal_type.TextToSpeechAudioPacket{
							ContextID:  contextID,
							AudioChunk: rawAudioData,
						})
					} else {
						st.logger.Errorf("smallest-tts: error decoding base64 audio: %v", err)
					}
				}
			case "complete":
				st.onPacket(internal_type.TextToSpeechEndPacket{
					ContextID: contextID,
				})
			case "error":
				st.logger.Errorf("smallest-tts: server error: %s (raw: %s)", audioData.Message, string(audioChunk))
			}
		}
	}
}

func (t *smallestTTS) Transform(ctx context.Context, in internal_type.LLMPacket) error {
	t.mu.Lock()
	cnn := t.connection
	t.currentContextID = in.ContextId()
	t.mu.Unlock()

	if cnn == nil {
		return fmt.Errorf("smallest-tts: websocket connection is not initialized")
	}

	voiceID := "ashley"
	if voiceIDValue, err := t.mdlOpts.GetString("speak.voice.id"); err == nil {
		voiceID = voiceIDValue
	}

	language := "en"
	if langValue, err := t.mdlOpts.GetString("speak.language"); err == nil {
		language = langValue
	}

	switch input := in.(type) {
	case internal_type.LLMStreamPacket:
		payload := map[string]interface{}{
			"text":        input.Text,
			"voice_id":    voiceID,
			"language":    language,
			"sample_rate": int(t.GetSampleRate()),
			"speed":       1.0,
			"continue":    true,
			"flush":       true,
		}
		t.logger.Infof("smallest-tts: sending text chunk voice_id=%s language=%s sample_rate=%d len=%d", voiceID, language, t.GetSampleRate(), len(input.Text))

		if err := cnn.WriteJSON(payload); err != nil {
			t.logger.Errorf("smallest-tts: unable to write json for text to speech: %v", err)
			return err
		}
	case internal_type.LLMMessagePacket:
		// Send flush to signal end of stream
		payload := map[string]interface{}{
			"text":        "",
			"voice_id":    voiceID,
			"language":    language,
			"sample_rate": int(t.GetSampleRate()),
			"speed":       1.0,
			"continue":    false,
			"flush":       true,
		}
		t.logger.Infof("smallest-tts: sending flush (end of stream)")
		if err := cnn.WriteJSON(payload); err != nil {
			t.logger.Errorf("smallest-tts: unable to write flush for text to speech: %v", err)
			return err
		}
	default:
		return fmt.Errorf("smallest-tts: unsupported input type %T", in)
	}
	return nil
}

func (t *smallestTTS) Close(ctx context.Context) error {
	t.ctxCancel()
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connection != nil {
		t.connection.Close()
		t.connection = nil
	}
	return nil
}

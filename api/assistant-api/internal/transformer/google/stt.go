// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_google

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
)

type googleSpeechToText struct {
	*googleOption
	mu sync.Mutex

	logger commons.Logger

	client   *speech.Client
	stream   speechpb.Speech_StreamingRecognizeClient
	onPacket func(pkt ...internal_type.Packet) error

	// context management
	ctx       context.Context
	ctxCancel context.CancelFunc
}

// Name implements internal_transformer.SpeechToTextTransformer.
func (g *googleSpeechToText) Name() string {
	return "google-speech-to-text"
}

func NewGoogleSpeechToText(ctx context.Context, logger commons.Logger, credential *protos.VaultCredential, audioConfig *protos.AudioConfig,
	onPacket func(pkt ...internal_type.Packet) error,
	opts utils.Option,
) (internal_type.SpeechToTextTransformer, error) {
	start := time.Now()
	googleOption, err := NewGoogleOption(logger, credential, audioConfig, opts)
	if err != nil {
		logger.Errorf("google-stt: Error while GoogleOption err: %v", err)
		return nil, err
	}
	client, err := speech.NewClient(ctx, googleOption.GetSpeechToTextClientOptions()...)

	if err != nil {
		logger.Errorf("google-stt: Error creating Google client: %v", err)
		return nil, err
	}

	xctx, contextCancel := context.WithCancel(ctx)
	// Context for callback management
	logger.Benchmark("google.NewGoogleSpeechToText", time.Since(start))
	return &googleSpeechToText{
		ctx:          xctx,
		ctxCancel:    contextCancel,
		logger:       logger,
		client:       client,
		googleOption: googleOption,
		onPacket:     onPacket,
	}, nil
}

// Transform implements internal_transformer.SpeechToTextTransformer.
func (google *googleSpeechToText) Transform(c context.Context, in internal_type.UserAudioPacket) error {
	google.mu.Lock()
	strm := google.stream
	google.mu.Unlock()

	if strm == nil {
		return fmt.Errorf("google-stt: stream not initialized")
	}

	return strm.Send(&speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
			Audio: in.Audio,
		},
	})
}

// speechToTextCallback processes streaming responses with context awareness.
func (g *googleSpeechToText) speechToTextCallback(stram speechpb.Speech_StreamingRecognizeClient, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			g.logger.Infof("google-stt: context cancelled, stopping response listener")
			return
		default:
			resp, err := stram.Recv()
			if err == io.EOF {
				g.logger.Infof("google-stt: stream ended (EOF)")
				return
			}
			if err != nil {
				g.logger.Errorf("google-stt: recv error: %v", err)
				return
			}
			if resp == nil {
				g.logger.Warnf("google-stt: received nil response")
				return
			}

			for _, result := range resp.Results {
				if len(result.Alternatives) == 0 {
					continue
				}
				alt := result.Alternatives[0]
				if g.onPacket != nil && len(alt.GetTranscript()) > 0 {
					if v, err := g.mdlOpts.GetFloat64("listen.threshold"); err == nil {
						if alt.GetConfidence() < float32(v) {
							g.onPacket(
								internal_type.SpeechToTextPacket{
									Script:     alt.GetTranscript(),
									Confidence: float64(alt.GetConfidence()),
									Language:   result.GetLanguageCode(),
									Interim:    true,
								},
							)
							continue
						}
					}
					g.onPacket(
						internal_type.InterruptionPacket{Source: internal_type.InterruptionSourceWord},
						internal_type.SpeechToTextPacket{
							Script:     alt.GetTranscript(),
							Confidence: float64(alt.GetConfidence()),
							Language:   result.GetLanguageCode(),
							Interim:    !result.GetIsFinal(),
						},
					)
				}
			}

		}
	}
}

func (google *googleSpeechToText) Initialize() error {

	stream, err := google.client.StreamingRecognize(google.ctx)
	if err != nil {
		google.logger.Errorf("google-stt: error creating google-stt stream: %v", err)
		return err
	}

	if google.stream != nil {
		_ = google.stream.CloseSend()
	}

	google.mu.Lock()
	google.stream = stream
	defer google.mu.Unlock()

	if err := google.stream.Send(&speechpb.StreamingRecognizeRequest{
		Recognizer: google.GetRecognizer(),
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: google.SpeechToTextOptions(),
		},
	}); err != nil {
		google.logger.Errorf("google-stt: error creating google-stt stream: %v", err)
		return err
	}
	// Launch callback listener
	go google.speechToTextCallback(stream, google.ctx)
	google.logger.Debugf("google-stt: connection established")
	return nil
}

func (g *googleSpeechToText) Close(ctx context.Context) error {
	g.ctxCancel()

	g.mu.Lock()
	defer g.mu.Unlock()

	var combinedErr error
	if g.stream != nil {
		if err := g.stream.CloseSend(); err != nil {
			combinedErr = fmt.Errorf("error closing StreamClient: %v", err)
			g.logger.Errorf(combinedErr.Error())
		}
	}

	if g.client != nil {
		if err := g.client.Close(); err != nil {
			// Log the error if closure fails.
			combinedErr = fmt.Errorf("error closing Client: %v", err)
			g.logger.Errorf(combinedErr.Error())
		}
	}
	return combinedErr
}

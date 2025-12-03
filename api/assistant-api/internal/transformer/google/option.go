// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_google

import (
	"cloud.google.com/go/speech/apiv1/speechpb"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	internal_audio "github.com/rapidaai/api/assistant-api/internal/audio"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
	protos "github.com/rapidaai/protos"
	"google.golang.org/api/option"
)

type GoogleOption interface {
	SpeechToTextOptions() *speechpb.StreamingRecognitionConfig
	TextToSpeechOptions() *texttospeechpb.StreamingSynthesizeConfig
	GetClientOptions() []option.ClientOption
}

type googleOption struct {
	logger commons.Logger

	//
	clientOptons      []option.ClientOption
	audioConfig       *internal_audio.AudioConfig
	initializeOptions utils.Option
}

func GetSpeechToTextEncodingFromString(encoding internal_audio.AudioFormat) speechpb.RecognitionConfig_AudioEncoding {
	switch encoding {
	case internal_audio.Linear16:
		return speechpb.RecognitionConfig_LINEAR16
	case internal_audio.MuLaw8:
		return speechpb.RecognitionConfig_MULAW
	default:
		return speechpb.RecognitionConfig_LINEAR16
	}
}

func NewGoogleOption(logger commons.Logger,
	vaultCredential *protos.VaultCredential,
	audioConfig *internal_audio.AudioConfig,
	opts utils.Option) (GoogleOption, error) {
	cx, ok := vaultCredential.GetValue().AsMap()["key"]
	co := make([]option.ClientOption, 0)
	if ok {
		co = append(co, option.WithAPIKey(cx.(string)))
	}
	prj, ok := vaultCredential.GetValue().AsMap()["project_id"]
	if ok {
		co = append(co, option.WithQuotaProject(prj.(string)))
	}
	serviceCrd, ok := vaultCredential.GetValue().AsMap()["service_account_key"]
	if ok {
		serviceCrdJSON := []byte(serviceCrd.(string)) // Convert string to []byte
		co = append(co, option.WithCredentialsJSON(serviceCrdJSON))
	}

	return &googleOption{
		logger:            logger,
		initializeOptions: opts,
		clientOptons:      co,
		audioConfig:       audioConfig,
	}, nil
}

// GetCredential returns the credential string if present in opts, otherwise returns an empty string.
func (gO *googleOption) GetClientOptions() []option.ClientOption {
	return gO.clientOptons
}

func (gog *googleOption) SpeechToTextOptions() *speechpb.StreamingRecognitionConfig {
	opts := &speechpb.StreamingRecognitionConfig{
		Config: &speechpb.RecognitionConfig{
			Encoding:                   GetSpeechToTextEncodingFromString(gog.audioConfig.Format),
			SampleRateHertz:            int32(gog.audioConfig.GetSampleRate()),
			EnableAutomaticPunctuation: true,
			EnableWordConfidence:       true,
			ProfanityFilter:            true,
			LanguageCode:               "en-US",
			Model:                      "default",
			UseEnhanced:                true,
		},
		InterimResults: true,
	}

	if language, err := gog.initializeOptions.GetString("listen.language"); err == nil {
		opts.Config.LanguageCode = language
	}
	gog.logger.Debugf("language ============== %s", opts.Config.LanguageCode)
	if model, err := gog.initializeOptions.GetString("listen.model"); err == nil {
		opts.Config.Model = model
	}
	gog.logger.Debugf("model ============== %s", opts.Config.Model)

	return opts
}

func (goog *googleOption) TextToSpeechOptions() *texttospeechpb.StreamingSynthesizeConfig {
	options := &texttospeechpb.StreamingSynthesizeConfig{
		Voice: &texttospeechpb.VoiceSelectionParams{
			Name: "en-US-Chirp-HD-F",
		},
		StreamingAudioConfig: &texttospeechpb.StreamingAudioConfig{
			AudioEncoding:   GetTextToSpeechEncodingByName(goog.audioConfig.GetFormat()),
			SampleRateHertz: int32(goog.audioConfig.GetSampleRate()),
		},
	}

	voice, err := goog.initializeOptions.GetString("speak.voice.id")
	if err != nil {
		voice = "en-US-Chirp-HD-F"
	}

	options.Voice.Name = voice
	if sampleRate, err := goog.initializeOptions.GetUint32("speak.output_format.sample_rate"); err == nil {
		options.StreamingAudioConfig.SampleRateHertz = int32(sampleRate)
	}

	if encoding, err := goog.initializeOptions.GetString("speak.output_format.encoding"); err == nil {
		options.StreamingAudioConfig.AudioEncoding = GetTextToSpeechEncodingByName(encoding)
	}

	return options

}

func GetTextToSpeechEncodingByName(name string) texttospeechpb.AudioEncoding {
	switch name {
	case "AUDIO_ENCODING_UNSPECIFIED":
		return texttospeechpb.AudioEncoding_AUDIO_ENCODING_UNSPECIFIED
	case "MP3":
		return texttospeechpb.AudioEncoding_MP3
	case "OGG_OPUS":
		return texttospeechpb.AudioEncoding_OGG_OPUS
	case "MULAW", "MuLaw8":
		return texttospeechpb.AudioEncoding_MULAW
	case "ALAW":
		return texttospeechpb.AudioEncoding_ALAW
	case "PCM", "Linear16":
		return texttospeechpb.AudioEncoding_PCM
	case "M4A":
		return texttospeechpb.AudioEncoding_M4A
	default:
		return texttospeechpb.AudioEncoding_AUDIO_ENCODING_UNSPECIFIED
	}
}

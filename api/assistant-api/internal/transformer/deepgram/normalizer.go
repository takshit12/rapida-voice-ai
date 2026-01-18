// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.

package internal_transformer_deepgram

import (
	"context"
	"regexp"
	"strings"

	internal_normalizers "github.com/rapidaai/api/assistant-api/internal/normalizers"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

// =============================================================================
// Deepgram Text Normalizer
// =============================================================================

// deepgramNormalizer handles Deepgram TTS text preprocessing.
// Deepgram does NOT support SSML - only plain text is accepted.
type deepgramNormalizer struct {
	logger   commons.Logger
	config   internal_type.NormalizerConfig
	language string

	// normalizer pipeline
	normalizers []internal_normalizers.Normalizer
}

// NewDeepgramNormalizer creates a Deepgram-specific text normalizer.
func NewDeepgramNormalizer(logger commons.Logger, opts utils.Option) internal_type.TextNormalizer {
	cfg := internal_type.DefaultNormalizerConfig()

	logger.Debugf("testing -> deepgram-normalizer: initializing with options: %v", opts)
	language, _ := opts.GetString("speaker.language")
	if language == "" {
		language = "en"
	}

	// Build normalizer pipeline based on speaker.pronunciation.dictionaries
	var normalizers []internal_normalizers.Normalizer
	if dictionaries, err := opts.GetString("speaker.pronunciation.dictionaries"); err == nil && dictionaries != "" {
		normalizerNames := strings.Split(dictionaries, commons.SEPARATOR)
		normalizers = internal_type.BuildNormalizerPipeline(logger, normalizerNames)
	}

	return &deepgramNormalizer{
		logger:      logger,
		config:      cfg,
		language:    language,
		normalizers: normalizers,
	}
}

// Normalize applies Deepgram-specific text transformations.
// Deepgram does NOT support SSML, so we only normalize text without XML escaping.
func (n *deepgramNormalizer) Normalize(ctx context.Context, text string) string {
	if text == "" {
		return text
	}

	// Clean markdown first
	text = n.removeMarkdown(text)

	// Apply normalizer pipeline
	for _, normalizer := range n.normalizers {
		text = normalizer.Normalize(text)
	}

	// NO XML escaping - Deepgram uses plain text only
	// NO SSML breaks - Deepgram doesn't support SSML

	return n.normalizeWhitespace(text)
}

// =============================================================================
// Private Helpers
// =============================================================================

func (n *deepgramNormalizer) removeMarkdown(input string) string {
	re := regexp.MustCompile(`(?m)^#{1,6}\s*`)
	output := re.ReplaceAllString(input, "")

	re = regexp.MustCompile(`\*{1,2}([^*]+?)\*{1,2}|_{1,2}([^_]+?)_{1,2}`)
	output = re.ReplaceAllString(output, "$1$2")

	re = regexp.MustCompile("`([^`]+)`")
	output = re.ReplaceAllString(output, "$1")

	re = regexp.MustCompile("(?s)```[^`]*```")
	output = re.ReplaceAllString(output, "")

	re = regexp.MustCompile(`(?m)^>\s?`)
	output = re.ReplaceAllString(output, "")

	re = regexp.MustCompile(`\[(.*?)\]\(.*?\)`)
	output = re.ReplaceAllString(output, "$1")

	re = regexp.MustCompile(`!\[(.*?)\]\(.*?\)`)
	output = re.ReplaceAllString(output, "$1")

	re = regexp.MustCompile(`(?m)^(-{3,}|\*{3,}|_{3,})$`)
	output = re.ReplaceAllString(output, "")

	re = regexp.MustCompile(`[*_]+`)
	output = re.ReplaceAllString(output, "")

	return output
}

func (n *deepgramNormalizer) normalizeWhitespace(text string) string {
	re := regexp.MustCompile(`\s+`)
	result := re.ReplaceAllString(text, " ")
	return strings.TrimSpace(result)
}

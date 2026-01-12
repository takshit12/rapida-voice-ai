// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_sentence_tokenizer

import (
	internal_tokenizer "github.com/rapidaai/api/assistant-api/internal/tokenizer"
	internal_default "github.com/rapidaai/api/assistant-api/internal/tokenizer/sentence/internal/default"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

type TokenizerType string

const (
	TokenizerDefault TokenizerType = "default"
	TokenizerSimple  TokenizerType = "simple"
)

func NewSentenceTokenizer(
	logger commons.Logger,
	options utils.Option,
) (internal_tokenizer.SentenceTokenizer, error) {

	typ, _ := options.GetString("speaker.sentence.tokenizer")
	if typ == "" {
		typ = string(TokenizerDefault)
	}

	switch TokenizerType(typ) {
	default:
		return internal_default.NewSentenceTokenizer(logger, options)
	}
}

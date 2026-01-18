// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_sentence_assembler

import (
	"context"

	internal_default "github.com/rapidaai/api/assistant-api/internal/assembler/sentence/internal/default"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/utils"
)

type SentenceAssemblerType string

const (
	SentenceAssemblerDefault    SentenceAssemblerType = "default"
	OptionsKeySentenceAssembler string                = "speaker.sentence.assembler"
)

func GetLLMSentenceAssembler(
	context context.Context,
	logger commons.Logger,
	options utils.Option,
) (internal_type.LLMSentenceAssembler, error) {
	typ, _ := options.GetString(OptionsKeySentenceAssembler)
	switch SentenceAssemblerType(typ) {
	default:
		return internal_default.NewDefaultLLMSentenceAssembler(context, logger, options)
	}
}

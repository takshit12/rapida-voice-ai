// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"fmt"

	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	"github.com/rapidaai/protos"
)

func (io *GenericRequestor) Input(message *protos.AssistantConversationUserMessage) error {
	switch msg := message.GetMessage().(type) {
	case *protos.AssistantConversationUserMessage_Audio:
		return io.OnPacket(io.Context(), internal_type.UserAudioPacket{Audio: msg.Audio.GetContent()})
	case *protos.AssistantConversationUserMessage_Text:
		return io.OnPacket(io.Context(), internal_type.UserTextPacket{Text: msg.Text.GetContent()})

	default:
		return fmt.Errorf("illegal input from the user %+v", msg)
	}

}

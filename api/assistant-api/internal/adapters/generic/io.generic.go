// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package internal_adapter_generic

import (
	"context"
	"fmt"

	internal_adapter_request_customizers "github.com/rapidaai/api/assistant-api/internal/adapters/customizers"
	internal_type "github.com/rapidaai/api/assistant-api/internal/type"
	type_enums "github.com/rapidaai/pkg/types/enums"
	"github.com/rapidaai/pkg/utils"
	"github.com/rapidaai/protos"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (io *GenericRequestor) Input(message *protos.AssistantConversationUserMessage) error {
	switch msg := message.GetMessage().(type) {
	case *protos.AssistantConversationUserMessage_Audio:
		if v, err := io.ListenAudio(io.Context(), msg.Audio.GetContent()); err == nil {
			utils.Go(context.Background(), func() {
				io.recorder.User(v)
			})
		}
		return nil
	case *protos.AssistantConversationUserMessage_Text:
		io.messaging.Transition(internal_adapter_request_customizers.Interrupted)
		interim := io.messaging.Create(type_enums.UserActor, msg.Text.GetContent())
		if err := io.Notify(io.Context(), &protos.AssistantConversationUserMessage{Id: interim.GetId(), Completed: false, Message: &protos.AssistantConversationUserMessage_Text{Text: &protos.AssistantConversationMessageTextContent{Content: interim.String()}}, Time: timestamppb.Now()}); err != nil {
			io.logger.Tracef(io.Context(), "error while notifying the text input from user: %w", err)
		}
		return io.OnPacket(io.Context(), internal_type.UserTextPacket{Text: interim.String()})

	default:
		return fmt.Errorf("illegal input from the user %+v", msg)
	}

}

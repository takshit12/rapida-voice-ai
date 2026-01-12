// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package utils

import "testing"

func TestRapidaStage_Get(t *testing.T) {
	tests := []struct {
		stage    RapidaStage
		expected string
	}{
		{AssistantConnectStage, "talk.assistant.connect"},
		{AssistantCreateConversationStage, "talk.assistant.connect.create-conversation"},
		{AssistantResumeConverstaionStage, "talk.assistant.connect.resume-conversation"},
		{AssistantListenConnectStage, "talk.assistant.listen.connect"},
		{AssistantSpeakConnectStage, "talk.assistant.speak.connect"},
		{AssistantListeningStage, "talk.assistant.listen.listening"},
		{AssistantUtteranceStage, "talk.assistant.utterance"},
		{AssistantInterruptStage, "talk.assistant.interrupt"},
		{AssistantAgentConnectStage, "talk.assistant.agent.connect"},
		{AssistantToolConnectStage, "talk.assistant.tool.connect"},
		{AssistantToolExecuteStage, "talk.assistant.tool.execute"},
		{AssistantAgentTextGenerationStage, "talk.assistant.agent.text-generation"},
		{AssistantSpeakingStage, "talk.assistant.speak.speaking"},
		{AssistantNotifyStage, "talk.assistant.notify"},
		{AssistantDisconnectStage, "talk.assistant.disconnect"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if result := tt.stage.Get(); result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

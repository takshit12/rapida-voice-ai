// Copyright (c) 2023-2025 RapidaAI
// Author: Prashant Srivastav <prashant@rapida.ai>
//
// Licensed under GPL-2.0 with Rapida Additional Terms.
// See LICENSE.md or contact sales@rapida.ai for commercial usage.
package utils

type RapidaStage string

const (
	AssistantConnectStage RapidaStage = "talk.assistant.connect"

	AssistantCreateConversationStage RapidaStage = "talk.assistant.connect.create-conversation"
	AssistantResumeConverstaionStage RapidaStage = "talk.assistant.connect.resume-conversation"

	AssistantListenConnectStage       RapidaStage = "talk.assistant.listen.connect"
	AssistantSpeakConnectStage        RapidaStage = "talk.assistant.speak.connect"
	AssistantListeningStage           RapidaStage = "talk.assistant.listen.listening"
	AssistantUtteranceStage           RapidaStage = "talk.assistant.utterance"
	AssistantInterruptStage           RapidaStage = "talk.assistant.interrupt"
	AssistantAgentConnectStage        RapidaStage = "talk.assistant.agent.connect"
	AssistantToolConnectStage         RapidaStage = "talk.assistant.tool.connect"
	AssistantToolExecuteStage         RapidaStage = "talk.assistant.tool.execute"
	AssistantAgentTextGenerationStage RapidaStage = "talk.assistant.agent.text-generation"
	AssistantSpeakingStage            RapidaStage = "talk.assistant.speak.speaking"
	AssistantNotifyStage              RapidaStage = "talk.assistant.notify"
	AssistantDisconnectStage          RapidaStage = "talk.assistant.disconnect"
)

// Get returns the string value of the RapidaStage
func (r RapidaStage) Get() string {
	return string(r)
}

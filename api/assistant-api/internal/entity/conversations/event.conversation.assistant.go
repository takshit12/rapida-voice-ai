package internal_conversation_gorm

import (
	gorm_model "github.com/rapidaai/pkg/models/gorm"
	gorm_types "github.com/rapidaai/pkg/models/gorm/types"
)

type AssistantConverstaionTelephonyEvent struct {
	gorm_model.Audited
	AssistantConversationId uint64                  `json:"assistantConversationId" gorm:"type:bigint;not null"`
	EventType               string                  `json:"event_type" gorm:"type:string;size:200;not null"`
	Payload                 gorm_types.InterfaceMap `json:"payload" gorm:"type:string;size:200;not null"`
}

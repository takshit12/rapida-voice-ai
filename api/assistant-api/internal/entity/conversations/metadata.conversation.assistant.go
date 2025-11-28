package internal_conversation_gorm

import gorm "github.com/rapidaai/pkg/models/gorm"

type AssistantConversationMetadata struct {
	gorm.Audited
	gorm.Mutable
	gorm.Metadata
	AssistantConversationId uint64 `json:"assistantConversationId" gorm:"type:bigint;not null"`
}

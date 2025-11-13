package internal_entity

import gorm_model "github.com/rapidaai/pkg/models/gorm"

type NotificationSetting struct {
	gorm_model.Audited
	gorm_model.Mutable
	UserAuthId uint64 `json:"userAuthId" gorm:"type:bigint;size:20;not null;"`
	EventType  string `json:"event_type" gorm:"type:string;size:200;not null;"`
	Channel    string `json:"channel" gorm:"type:string;size:200;not null"`
	Enabled    bool   `json:"enabled" gorm:"type:bool;not null;default:false"`
}

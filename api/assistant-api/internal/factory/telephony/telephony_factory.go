package internal_telephony_factory

import (
	"errors"

	"github.com/rapidaai/api/assistant-api/config"
	internal_telephony "github.com/rapidaai/api/assistant-api/internal/telephony"
	internal_exotel_telephony "github.com/rapidaai/api/assistant-api/internal/telephony/exotel"
	internal_twilio_telephony "github.com/rapidaai/api/assistant-api/internal/telephony/twilio"
	internal_vonage_telephony "github.com/rapidaai/api/assistant-api/internal/telephony/vonage"
	"github.com/rapidaai/pkg/commons"
)

type Telephony string

const (
	Twilio Telephony = "twilio"
	Exotel Telephony = "exotel"
	Vonage Telephony = "vonage"
)

func (at Telephony) String() string {
	return string(at)
}

func GetTelephony(
	at Telephony,
	cfg *config.AssistantConfig,
	logger commons.Logger) (internal_telephony.Telephony, error) {
	switch at {
	case Twilio:
		return internal_twilio_telephony.NewTwilioTelephony(cfg, logger)
	case Exotel:
		return internal_exotel_telephony.NewExotelTelephony(cfg, logger)
	case Vonage:
		return internal_vonage_telephony.NewVonageTelephony(cfg, logger)
	default:
		return nil, errors.New("illegal telephony provider")
	}
}

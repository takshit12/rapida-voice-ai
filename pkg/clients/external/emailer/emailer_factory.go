package external_emailer

import (
	"github.com/rapidaai/config"
	external_clients "github.com/rapidaai/pkg/clients/external"
	"github.com/rapidaai/pkg/commons"
	"github.com/rapidaai/pkg/configs"
)

func NewEmailer(config *config.AppConfig, logger commons.Logger) external_clients.Emailer {
	if config.EmailerConfig == nil {
		return NewLocalEmailer(logger)
	}
	switch config.EmailerConfig.Provider() {
	case configs.SENDGRID:
		return NewSendgridEmailer(logger, config.EmailerConfig)
	case configs.SES:
		return NewSESEmailer(logger, config.EmailerConfig)
	default:
		return NewLocalEmailer(logger)
	}
}

package configs

type EmailProvider string

const (
	SES      EmailProvider = "ses"
	SENDGRID EmailProvider = "sendgrid"
)

// asset_upload_bucket

type EmailerConfig struct {
	EmailProvider string     `mapstructure:"provider" validate:"required"`
	FromEmail     string     `mapstructure:"from_email" validate:"required"`
	FromName      string     `mapstructure:"from_name" validate:"required"`
	Auth          *AwsConfig `mapstructure:"auth"`
	SendgridKey   *string    `mapstructure:"sendgrid_key"`
}

func (cfg *EmailerConfig) Provider() EmailProvider {
	if cfg.EmailProvider == string(SES) {
		return SES
	}
	if cfg.EmailProvider == string(SENDGRID) {
		return SENDGRID
	}
	return SENDGRID
}

package config

import "github.com/rapidaai/pkg/configs"

type AppConfig struct {
	//
	Name     string `mapstructure:"service_name" validate:"required"`
	Version  string `mapstructure:"version"`
	Host     string `mapstructure:"host" validate:"required"`
	Env      string `mapstructure:"env" validate:"required"`
	Port     int    `mapstructure:"port" validate:"required"`
	LogLevel string `mapstructure:"log_level" validate:"required"`
	Secret   string `mapstructure:"secret" validate:"required"`

	// all the host
	IntegrationHost string `mapstructure:"integration_host" validate:"required"`
	EndpointHost    string `mapstructure:"endpoint_host" validate:"required"`
	AssistantHost   string `mapstructure:"assistant_host" validate:"required"`
	WebHost         string `mapstructure:"web_host" validate:"required"`
	DocumentHost    string `mapstructure:"document_host" validate:"required"`

	// utility
	UiHost        string                 `mapstructure:"ui_host" validate:"required"`
	EmailerConfig *configs.EmailerConfig `mapstructure:"emailer"`
}

func (cfg *AppConfig) IsDevelopment() bool {
	return cfg.Env != "production"
}

func (cfg *AppConfig) BaseUrl() (baseUrl string) {
	return cfg.UiHost
}

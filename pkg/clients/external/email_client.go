package external_clients

import (
	"context"

	external_emailer_template "github.com/rapidaai/pkg/clients/external/emailer/template"
)

type Contact struct {
	Name  string
	Email string
}

type Emailer interface {
	EmailText(ctx context.Context, to Contact, subject string, content string) error
	EmailRichText(ctx context.Context, to Contact, subject string, template external_emailer_template.TemplateName, args map[string]string) error
	EmailTemplate(ctx context.Context, to Contact, subject string, templateId string, args map[string]string) error
}

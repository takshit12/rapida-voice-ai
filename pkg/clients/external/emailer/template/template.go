package external_emailer_template

import (
	"embed"
	"html/template"
	"strings"
	texttpl "text/template"
)

type TemplateName string

var (
	INVITE_MEMBER_TEMPLATE      TemplateName = "invite-member.email"
	RESET_PASSWORD_TEMPLATE     TemplateName = "reset-password.email"
	WELCOME_MEMBER_TEMPLATE     TemplateName = "welcome.email"
	EMAIL_VERIFICATION_TEMPLATE TemplateName = "email-verification.email"
	NOTIFICATION_TEMPLATE       TemplateName = "notification.email"
)

//go:embed *.txt *.html
var templatesFS embed.FS

// GetTextTemplate returns a parsed text/template (for plain-text templates)
func GetTextTemplate(name TemplateName) (*texttpl.Template, error) {
	tmpl := string(name)
	if !strings.HasSuffix(tmpl, ".txt") {
		tmpl = tmpl + ".txt"
	}
	return texttpl.ParseFS(templatesFS, tmpl)
}

// GetHTMLTemplate returns a parsed html/template (for HTML templates)
func GetHTMLTemplate(name TemplateName) (*template.Template, error) {
	tmpl := string(name)
	if !strings.HasSuffix(tmpl, ".html") {
		tmpl = tmpl + ".html"
	}
	return template.ParseFS(templatesFS, tmpl)
}

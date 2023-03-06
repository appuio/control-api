package mailsenders

import (
	"context"

	"github.com/mailgun/mailgun-go/v4"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type MailSender interface {
	Send(ctx context.Context, recipient string, invitation string, token string) (string, error)
}

type MailgunSender struct {
	Mailgun        *mailgun.MailgunImpl
	MailgunBaseUrl string
	SenderAddress  string
	TemplateName   string
	UseTestMode    bool
	Subject        string
}

type LogSender struct{}

func (s *LogSender) Send(ctx context.Context, recipient string, invitation string, token string) (string, error) {
	log := log.FromContext(ctx)
	log.V(0).Info("E-mail backend is 'stdout'; invitation e-mail was not sent", "recipient", recipient, "invitation", invitation)
	return "", nil
}

func NewMailgunSender(domain string, token string, baseUrl string, senderAddress string, templateName string, subject string, useTestMode bool) MailgunSender {
	mg := mailgun.NewMailgun(domain, token)
	mg.SetAPIBase(baseUrl)
	return MailgunSender{
		Mailgun:       mg,
		SenderAddress: senderAddress,
		TemplateName:  templateName,
		UseTestMode:   useTestMode,
		Subject:       subject,
	}
}

func (m *MailgunSender) Send(ctx context.Context, recipient string, invitation string, token string) (string, error) {
	message := m.Mailgun.NewMessage(
		m.SenderAddress,
		m.Subject,
		"", // Message body will be rendered from template
		recipient,
	)
	message.SetTemplate(m.TemplateName)
	err := message.AddTemplateVariable("invitation", invitation)
	if err != nil {
		return "", err
	}
	err = message.AddTemplateVariable("token", token)
	if err != nil {
		return "", err
	}

	if m.UseTestMode {
		message.EnableTestMode()
	}

	_, id, err := m.Mailgun.Send(ctx, message)
	return id, err
}

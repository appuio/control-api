package mailsenders

import (
	"context"

	"github.com/mailgun/mailgun-go/v4"
	"sigs.k8s.io/controller-runtime/pkg/log"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

type MailSender interface {
	Send(ctx context.Context, recipient string, inv userv1.Invitation) (id string, err error)
}

var _ MailSender = &MailgunSender{}

// MailgunSender is a MailSender that sends e-mails via Mailgun.
type MailgunSender struct {
	Mailgun        mailgun.Mailgun
	MailgunBaseUrl string
	SenderAddress  string
	UseTestMode    bool
	Subject        string

	Body *InvitationRenderer
}

var _ MailSender = &LogSender{}

// LogSender is a MailSender that logs the e-mail to stdout.
type LogSender struct {
	Body *InvitationRenderer
}

func (s *LogSender) Send(ctx context.Context, recipient string, inv userv1.Invitation) (string, error) {
	log := log.FromContext(ctx)

	body, err := s.Body.Render(inv)
	if err != nil {
		return "", err
	}

	log.V(0).Info("E-mail body", "body", body)
	log.V(0).Info("E-mail backend is 'stdout'; invitation e-mail was not sent", "recipient", recipient)

	return "", nil
}

func NewMailgunSender(domain string, token string, baseUrl string, senderAddress string, body *InvitationRenderer, subject string, useTestMode bool) MailgunSender {
	mg := mailgun.NewMailgun(domain, token)
	mg.SetAPIBase(baseUrl)
	return MailgunSender{
		Mailgun:       mg,
		Body:          body,
		SenderAddress: senderAddress,
		UseTestMode:   useTestMode,
		Subject:       subject,
	}
}

func (m *MailgunSender) Send(ctx context.Context, recipient string, inv userv1.Invitation) (string, error) {
	body, err := m.Body.Render(inv)
	if err != nil {
		return "", err
	}

	message := m.Mailgun.NewMessage(
		m.SenderAddress,
		m.Subject,
		body,
		recipient,
	)

	if m.UseTestMode {
		message.EnableTestMode()
	}

	_, id, err := m.Mailgun.Send(ctx, message)
	return id, err
}

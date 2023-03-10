package mailsenders

import (
	"fmt"
	"strings"
	"text/template"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

type InvitationRenderer struct {
	Template *template.Template
}

type templateData struct {
	Invitation userv1.Invitation
}

func (ir InvitationRenderer) Render(inv userv1.Invitation) (string, error) {
	body := &strings.Builder{}
	err := ir.Template.Execute(body, templateData{Invitation: inv})
	if err != nil {
		return "", fmt.Errorf("failed to render e-mail body: %w", err)
	}
	return body.String(), nil
}

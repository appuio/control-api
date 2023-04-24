package mailsenders_test

import (
	"testing"
	"text/template"

	"github.com/appuio/control-api/mailsenders"
	"github.com/stretchr/testify/assert"

	userv1 "github.com/appuio/control-api/apis/user/v1"
)

func Test_InvitationRenderer_Render(t *testing.T) {
	tm, err := template.New("test").Parse("Hi {{.Object.Spec.Email}}, get your token: {{.Object.Status.Token}}")
	assert.NoError(t, err)

	subject := mailsenders.Renderer{Template: tm}
	rendered, err := subject.Render(userv1.Invitation{
		Spec: userv1.InvitationSpec{
			Email: "test@example.com",
		},
		Status: userv1.InvitationStatus{
			Token: "abc",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "Hi test@example.com, get your token: abc", rendered)
}

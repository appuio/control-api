package mailsenders

import (
	"fmt"
	"strings"
	"text/template"
)

type Renderer struct {
	Template *template.Template
}

type TemplateData struct {
	Object any
}

func (ir Renderer) Render(obj any) (string, error) {
	body := &strings.Builder{}
	err := ir.Template.Execute(body, TemplateData{Object: obj})
	if err != nil {
		return "", fmt.Errorf("failed to render e-mail body: %w", err)
	}
	return body.String(), nil
}

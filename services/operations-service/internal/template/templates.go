package template

import (
	"bytes"
	htmltemplate "html/template"
	texttemplate "text/template"
)

type Definition struct {
	Name    string
	Subject string
	HTML    string
	Text    string
}

var Definitions = map[string]Definition{
	"welcome":          {Name: "welcome", Subject: "Welcome to Logistics Platform", HTML: "<h1>Welcome {{.Name}}</h1><p>Verify: {{.VerificationLink}}</p>", Text: "Welcome {{.Name}}. Verify: {{.VerificationLink}}"},
	"magic_link":       {Name: "magic_link", Subject: "Your magic link", HTML: "<p>Login here: {{.Link}}</p>", Text: "Login here: {{.Link}}"},
	"password_reset":   {Name: "password_reset", Subject: "Reset your password", HTML: "<p>Reset: {{.Link}}</p>", Text: "Reset: {{.Link}}"},
	"order_confirmed":  {Name: "order_confirmed", Subject: "Order confirmed", HTML: "<p>Order {{.OrderID}} confirmed.</p>", Text: "Order {{.OrderID}} confirmed."},
	"order_assigned":   {Name: "order_assigned", Subject: "Driver assigned", HTML: "<p>{{.DriverName}} is assigned to order {{.OrderID}}.</p>", Text: "{{.DriverName}} is assigned to order {{.OrderID}}."},
	"order_delivered":  {Name: "order_delivered", Subject: "Order delivered", HTML: "<p>Order {{.OrderID}} delivered.</p>", Text: "Order {{.OrderID}} delivered."},
	"order_cancelled":  {Name: "order_cancelled", Subject: "Order cancelled", HTML: "<p>Order {{.OrderID}} cancelled: {{.Reason}}</p>", Text: "Order {{.OrderID}} cancelled: {{.Reason}}"},
	"payment_received": {Name: "payment_received", Subject: "Payment received", HTML: "<p>Payment received for order {{.OrderID}}.</p>", Text: "Payment received for order {{.OrderID}}."},
	"payment_failed":   {Name: "payment_failed", Subject: "Payment failed", HTML: "<p>Payment failed for order {{.OrderID}}. Retry: {{.RetryLink}}</p>", Text: "Payment failed for order {{.OrderID}}. Retry: {{.RetryLink}}"},
}

func Render(name string, data any) (subject, htmlBody, textBody string, err error) {
	definition, ok := Definitions[name]
	if !ok {
		return "", "", "", nil
	}
	htmlTmpl, err := htmltemplate.New(name).Parse(definition.HTML)
	if err != nil {
		return "", "", "", err
	}
	textTmpl, err := texttemplate.New(name).Parse(definition.Text)
	if err != nil {
		return "", "", "", err
	}
	var htmlBuf, textBuf bytes.Buffer
	if err := htmlTmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", "", err
	}
	if err := textTmpl.Execute(&textBuf, data); err != nil {
		return "", "", "", err
	}
	return definition.Subject, htmlBuf.String(), textBuf.String(), nil
}

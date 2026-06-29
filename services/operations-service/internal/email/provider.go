package email

import (
	"context"
	"fmt"
)

type Provider interface {
	Name() string
	Send(ctx context.Context, req Request) (Response, error)
}

type Request struct {
	To        string
	Subject   string
	BodyHTML  string
	BodyText  string
	FromName  string
	FromEmail string
}

type Response struct {
	ProviderMessageID string `json:"provider_message_id"`
	Status            string `json:"status"`
}

type MailtrapProvider struct{}
type ResendProvider struct{}

func (MailtrapProvider) Name() string { return "mailtrap" }
func (ResendProvider) Name() string   { return "resend" }

func (MailtrapProvider) Send(ctx context.Context, req Request) (Response, error) {
	if req.To == "" || req.Subject == "" {
		return Response{}, fmt.Errorf("mailtrap: invalid request")
	}
	return Response{ProviderMessageID: "mailtrap-" + req.To, Status: "sent"}, nil
}

func (ResendProvider) Send(ctx context.Context, req Request) (Response, error) {
	if req.To == "" || req.Subject == "" {
		return Response{}, fmt.Errorf("resend: invalid request")
	}
	return Response{ProviderMessageID: "resend-" + req.To, Status: "sent"}, nil
}

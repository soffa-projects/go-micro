package messaging

import "github.com/fabriqs/go-micro/h"

type EmailAddress struct {
	Name    string
	Address string
}

type Email struct {
	From         *EmailAddress
	To           []EmailAddress
	Subject      string
	Body         string
	TemplateId   string
	TemplateData h.Map
}

type EmailSender interface {
	Send(message Email) error
}

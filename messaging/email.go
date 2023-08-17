package messaging

import (
	"github.com/fabriqs/go-micro/util/h"
)

type EmailAddress struct {
	Name    string
	Address string
	Primary bool
}

type Email struct {
	From         *EmailAddress
	To           []EmailAddress
	Subject      string
	Body         string
	TemplateId   string
	TemplateData h.Map
}

type Mailer interface {
	Send(message Email) error
}

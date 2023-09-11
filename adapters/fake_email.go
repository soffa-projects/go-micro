package adapters

import (
	"github.com/fabriqs/go-micro/micro"
)

type FakeEmailSender struct {
	micro.Mailer
	EmailSent int
}

func NewFakeEmailSender() *FakeEmailSender {
	return &FakeEmailSender{
		EmailSent: 0,
	}
}

func (s *FakeEmailSender) Send(_ micro.Email) error {
	s.EmailSent++
	return nil
}

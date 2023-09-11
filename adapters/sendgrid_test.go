package adapters

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/fabriqs/go-micro/micro"
	"github.com/fabriqs/go-micro/util/h"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	sender := NewSendGridEmailSender("foo")
	err := sender.Send(micro.Email{
		From: &micro.EmailAddress{
			Name:    gofakeit.Name(),
			Address: gofakeit.Email(),
		},
		To: []micro.EmailAddress{{
			Name:    gofakeit.Name(),
			Address: gofakeit.Email(),
		}},
		Subject:    "Hello world",
		TemplateId: gofakeit.UUID(),
		TemplateData: h.Map{
			"email":      "hello@gmail.com",
			"name":       "Titus",
			"project":    "Test Project",
			"invite_url": "http://localhost:3000/invites/1234",
		},
	})
	assert.True(t, strings.Contains(err.Error(), "The provided authorization grant is invalid"))
	//assert.Nil(t, err)
}

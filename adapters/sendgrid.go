package adapters

import (
	"encoding/json"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/micro"
)

type SendGridEmailSender struct {
	micro.Mailer
	apiKey string
}

type sengridError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Help    string `json:"help"`
}

type sengridErrorResponse struct {
	Errors []sengridError `json:"errors"`
}

func NewSendGridEmailSender(apikey string) micro.Mailer {
	return SendGridEmailSender{apiKey: apikey}
}

func (s SendGridEmailSender) Send(message micro.Email) error {
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail(message.From.Name, message.From.Address))
	m.Subject = message.Subject
	p := mail.NewPersonalization()
	recipients := make([]string, 0)
	for _, to := range message.To {
		p.AddTos(mail.NewEmail(to.Name, to.Address))
		recipients = append(recipients, to.Address)
	}
	if message.TemplateId != "" {
		m.SetTemplateID(message.TemplateId)
		for k, v := range message.TemplateData {
			p.SetDynamicTemplateData(k, v)
		}
	}
	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(s.apiKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	res, err := sendgrid.API(request)
	if err != nil {
		log.Errorf(err.Error())
		return err
	} else if res.StatusCode >= 300 {
		var resultErr sengridErrorResponse
		_ = json.Unmarshal([]byte(res.Body), &resultErr)
		return fmt.Errorf("%s", resultErr.Errors[0].Message)
	} else {
		log.Infof("Email sent to %v", recipients)
		/*model := ProjectInvitation{Id: invite.Id}
		  if err = ctx.Env.DB.UpdateColumn(&model, "status", StatusSent); err != nil {
		    log.Error(err)
		  }*/
	}
	return nil
}

func (s SendGridEmailSender) SendBatch(message []micro.Email) error {
	for _, msg := range message {
		if err := s.Send(msg); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

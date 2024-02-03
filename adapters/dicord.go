package adapters

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/util/h"
)

type discordClient struct {
	micro.NotificationService
	webHookUrl string
	client     *resty.Client
}

func NewDiscordClient(webhook string) micro.NotificationService {
	if webhook == "" {
		log.Fatal("no notifications manager found, skipping notification")
		return nil
	}
	return &discordClient{
		webHookUrl: webhook,
		client:     resty.New(),
	}
}

func (s *discordClient) Send(_ micro.Ctx, message micro.Notification) error {
	out, err := s.client.R().
		SetBody(h.Map{
			"content": message.Message,
		}).
		SetContentLength(true).
		Post(s.webHookUrl)
	if err != nil {
		return fmt.Errorf("failed to send discord message -- %v", err)
	}
	if out.IsError() {
		return fmt.Errorf("failed to send discord message -- %v", out.Body())
	}
	return nil
}

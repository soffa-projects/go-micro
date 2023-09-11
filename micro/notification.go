package micro

import (
	log "github.com/sirupsen/logrus"
)

type Notification struct {
	Message string `json:"content"`
}

type NotificationService interface {
	Send(ctx Ctx, message Notification) error
}

type NoopNotificationService struct {
	Notification
}

func NewNoopNotificationService() *NoopNotificationService {
	return &NoopNotificationService{}
}

func (n *NoopNotificationService) Send(_ Ctx, message Notification) error {
	log.Warnf("no notifier configured -- message: %s", message.Message)
	return nil
}

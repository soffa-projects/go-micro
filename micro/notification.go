package micro

import (
	log "github.com/sirupsen/logrus"
)

type Notification struct {
	Message string `json:"content"`
}

type NotificationService interface {
	Send(message Notification) error
}

type NoopNotificationService struct {
	Notification
}

func NewNoopNotificationService() *NoopNotificationService {
	return &NoopNotificationService{}
}

func (n *NoopNotificationService) Send(message Notification) error {
	log.Warnf("no notifier configured -- message: %s", message.Message)
	return nil
}

package micro

import (
	"github.com/asaskevich/EventBus"
	"github.com/google/martian/v3/log"
)

var impl = EventBus.New()

type Event struct {
	Subject string
	Event   string
	Error   string
	Data    interface{}
}

type SubscribeFunc = func(ctx Ctx, payload Event) error

func Subscribe(topic string, handle SubscribeFunc) error {
	//ctx := micro.CurrentContext()
	return impl.Subscribe(topic, func(ctx Ctx, payload Event) {
		if err := handle(ctx, payload); err != nil {
			log.Errorf("error handling event: %s", err)
		}
	})
}

func SubscribeAsync(topic string, handle SubscribeFunc) error {
	//return impl.SubscribeAsync(topic, handle, false)
	return impl.SubscribeAsync(topic, func(ctx Ctx, payload Event) {
		if err := handle(ctx, payload); err != nil {
			log.Errorf("error handling event: %s", err)
		}
	}, false)
}

func SendNotification(ctx Ctx, event Notification) {
	Publish(ctx, NotificationTopic, Event{
		Event: event.Message,
	})
}

func Publish(ctx Ctx, topic string, payload Event) {
	if payload.Error != "" {
		log.Errorf(payload.Error)
	}
	impl.Publish(topic, ctx, payload)
}

func WaitAsync() {
	impl.WaitAsync()
}

func Reset() {
	impl.WaitAsync()
	impl = EventBus.New()
}

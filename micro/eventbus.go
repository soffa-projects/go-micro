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

type SubscribeFunc = func(payload Event) error

func Subscribe(topic string, handle SubscribeFunc) error {
	//ctx := micro.CurrentContext()
	return impl.Subscribe(topic, func(ctx Ctx, payload Event) {
		SetContext(ctx)
		if err := handle(payload); err != nil {
			log.Errorf("error handling event: %s", err)
		}
	})
}

func SubscribeAsync(topic string, handle SubscribeFunc) error {
	//return impl.SubscribeAsync(topic, handle, false)
	return impl.SubscribeAsync(topic, func(ctx Ctx, payload Event) {
		SetContext(ctx)
		if err := handle(payload); err != nil {
			log.Errorf("error handling event: %s", err)
		}
	}, false)
}

/*
func SubscribeOnceAsync(topic string, handle SubscribeFunc) error {
	return impl.SubscribeOnceAsync(topic, handle)
}

func Unsubscribe(topic string, handle SubscribeFunc) error {
	return impl.Unsubscribe(topic, handle)
}
*/

func SendNotification(event Notification) {
	Publish(NotificationTopic, Event{
		Event: event.Message,
	})
}

func Publish(topic string, payload Event) {
	if payload.Error != "" {
		log.Errorf(payload.Error)
	}
	ctx := CurrentContext()
	impl.Publish(topic, ctx, payload)
}

func WaitAsync() {
	impl.WaitAsync()
}

func Reset() {
	impl.WaitAsync()
	impl = EventBus.New()
}

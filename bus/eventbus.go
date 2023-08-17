package bus

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

func Subscribe(topic string, handle interface{}) error {
	return impl.Subscribe(topic, handle)
}

func SubscribeAsync(topic string, handle interface{}) error {
	return impl.SubscribeAsync(topic, handle, false)
}

func SubscribeOnceAsync(topic string, handle interface{}) error {
	return impl.SubscribeOnceAsync(topic, handle)
}

func Unsubscribe(topic string, handle interface{}) error {
	return impl.Unsubscribe(topic, handle)
}

func Publish(topic string, payload Event) {
	if payload.Error != "" {
		log.Errorf(payload.Error)
	}
	impl.Publish(topic, payload)
}

func WaitAsync() {
	impl.WaitAsync()
}

func Reset() {
	impl.WaitAsync()
	impl = EventBus.New()
}

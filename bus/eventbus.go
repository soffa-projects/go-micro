package bus

import (
	"github.com/asaskevich/EventBus"
	"github.com/fabriqs/go-micro/micro"
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
	return impl.Subscribe(topic, func(ctx micro.Ctx, payload Event) {
		micro.Invoke(ctx, func() (interface{}, error) {
			err := handle(payload)
			return nil, err
		})
	})
}

func SubscribeAsync(topic string, handle SubscribeFunc) error {
	//return impl.SubscribeAsync(topic, handle, false)
	return impl.SubscribeAsync(topic, func(ctx micro.Ctx, payload Event) {
		micro.Invoke(ctx, func() (interface{}, error) {
			err := handle(payload)
			return nil, err
		})
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

func Publish(topic string, payload Event) {
	if payload.Error != "" {
		log.Errorf(payload.Error)
	}
	ctx := micro.CurrentContext()
	impl.Publish(topic, ctx, payload)
}

func WaitAsync() {
	impl.WaitAsync()
}

func Reset() {
	impl.WaitAsync()
	impl = EventBus.New()
}

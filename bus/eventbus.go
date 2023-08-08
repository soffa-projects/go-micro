package bus

import "github.com/asaskevich/EventBus"

var impl = EventBus.New()

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

func Publish(topic string, message ...interface{}) {
	impl.Publish(topic, message...)
}

func WaitAsync() {
	impl.WaitAsync()
}

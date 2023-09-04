package micro

type Scheduler interface {
	StartAsync()
	Every(interval string, handler func() error)
	Once(handler func() error)
}

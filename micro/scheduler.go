package micro

type SchedulerHandler = func(ctx Ctx) error

type Scheduler interface {
	StartAsync()
	Every(interval string, handler SchedulerHandler)
	Once(handler SchedulerHandler)
	EveryTenant(interval string, handler SchedulerHandler)
	OncePerTenant(handler SchedulerHandler)
}

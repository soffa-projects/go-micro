package adapters

import (
	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/micro"
	"time"
)

type GoCronSchedulingAdapter struct {
	micro.Scheduler
	//ctx          micro.Ctx
	internal     *gocron.Scheduler
	tenantLoader micro.TenantLoader
}

func NewGoCronAdapter(tenantLoader micro.TenantLoader) micro.Scheduler {
	s := gocron.NewScheduler(time.UTC)
	return &GoCronSchedulingAdapter{
		internal:     s,
		tenantLoader: tenantLoader,
	}
}

func (s *GoCronSchedulingAdapter) StartAsync() {
	s.internal.StartAsync()
}

func (s *GoCronSchedulingAdapter) Every(interval string, handler micro.SchedulerHandler) {
	s.schedule(interval, 0, handler)
}

func (s *GoCronSchedulingAdapter) Once(handler micro.SchedulerHandler) {
	s.schedule("5s", 1, handler)
}

func (s *GoCronSchedulingAdapter) EveryTenant(interval string, handler micro.SchedulerHandler) {
	s.schedule(interval, 0, handler, s.tenantLoader.GetTenant()...)
}

func (s *GoCronSchedulingAdapter) OncePerTenant(handler micro.SchedulerHandler) {
	s.schedule("5s", 1, handler, s.tenantLoader.GetTenant()...)
}

func (s *GoCronSchedulingAdapter) schedule(interval string, limit int, handler func(ctx micro.Ctx) error, tenants ...string) {
	sched, err := s.internal.Every(interval).Do(func() error {
		defer func() {
			if err := recover(); err != nil {
				log.Error(err)
			}
		}()

		if tenants == nil || len(tenants) == 0 {
			err := handler(micro.NewCtx(micro.DefaultTenantId))
			if err != nil {
				log.Error(err)
			}
			return err

		} else {
			for _, tenantId := range tenants {
				err := handler(micro.NewCtx(tenantId))
				if err != nil {
					log.Error(err)
				}
			}
			return nil
		}
	})
	if limit > 0 {
		sched.LimitRunsTo(limit)
	}
	if err != nil {
		log.Fatal(err)
	}
}

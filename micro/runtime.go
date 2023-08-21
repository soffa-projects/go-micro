package micro

import (
	"github.com/fabriqs/go-micro/di"
	log "github.com/sirupsen/logrus"
	"github.com/timandy/routine"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func (a *App) Run(addr string) {
	// setup exit code for graceful shutdown
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// start the server
	go func() {
		_ = a.Router.Start("0.0.0.0:" + addr)
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = a.Router.Shutdown()
		if a.Env.DataSource != nil {
			a.Env.DataSource.Close()
		}
	}()

	if a.Scheduler != nil {
		go func() {
			time.Sleep(10 * time.Second)
			a.Scheduler.StartAsync()
			log.Infof("scheduler started")
		}()
	}
	/*if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}*/

	// ensure the server is shutdown gracefully & app runs
	gracefully()
}

func (a *App) Cleanup() {
	if a.Env.DataSource != nil {
		a.Env.DataSource.Close()
	}
}

func (a *App) Init() *App {
	//a.components = make([]Component, 0)

	if a.Scheduler != nil {
		di.Register(SchedulerService, a.Scheduler)
	}
	if a.TokenProvider != nil {
		di.Register(TokenProviderService, a.TokenProvider)
	}
	if a.Mailer != nil {
		di.Register(MailerServer, a.Mailer)
	}

	if a.Notifier != nil {
		di.Register(Notifications, a.Notifier)
		Subscribe(NotificationTopic, func(payload Event) error {
			return a.Notifier.Send(Notification{
				Message: payload.Event,
			})
		})
	}

	for _, feat := range a.Features {
		if feat.Init != nil {
			component, err := feat.Init(a)
			if err != nil {
				log.Fatalf("failed to init feature %s.\n%v", feat.Name, err)
				return nil
			}
			//a.components = append(a.components, component)
			if component != nil {
				di.Register(feat.Name, component)
			}
		}
	}
	return a
}

func gracefully() {
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func Invoke(ctx Ctx, cb func() (interface{}, error)) (out interface{}, err error) {
	SetContext(ctx)
	defer func() {
		ClearContext()
	}()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	routine.Go(func() {
		defer func() {
			if err0 := recover(); err0 != nil {
				err = err0.(error)
			}
			wg.Done()
		}()
		out, err = cb()
	})
	wg.Wait()
	return out, err
}

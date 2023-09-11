package micro

import (
	"github.com/fabriqs/go-micro/di"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *App) Run(addr string) {
	// setup exit code for graceful shutdown
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	// start the server
	go func() {
		_ = app.Env.Router.Start("0.0.0.0:" + addr)
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = app.Env.Router.Shutdown()
		if app.Env.DataSource != nil {
			app.Env.DataSource.Close()
		}
	}()

	if app.Env.Scheduler != nil {
		go func() {
			time.Sleep(10 * time.Second)
			app.Env.Scheduler.StartAsync()
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

func (app *App) Cleanup() {
	if app.Env.DataSource != nil {
		app.Env.DataSource.Close()
	}
}

func (app *App) Init() *App {
	//env.components = make([]Component, 0)

	env := app.Env

	globalEnv = env
	globalApp = app

	if env.Scheduler != nil {
		di.Register(SchedulerService, env.Scheduler)
	}
	if env.TokenProvider != nil {
		di.Register(TokenProviderService, env.TokenProvider)
	}
	if env.Mailer != nil {
		di.Register(MailerServer, env.Mailer)
	}

	if env.Notifier != nil {
		di.Register(Notifications, env.Notifier)
		Subscribe(NotificationTopic, func(ctx Ctx, payload Event) error {
			return env.Notifier.Send(ctx, Notification{
				Message: payload.Event,
			})
		})
	}

	for _, feat := range app.Features {
		if feat.Init != nil {
			component, err := feat.Init(app)
			if err != nil {
				log.Fatalf("failed to init feature %s.\n%v", feat.Name, err)
			}
			//env.components = append(env.components, component)
			if component != nil {
				di.Register(feat.Name, component)
			}
		}
	}

	return app
}

func gracefully() {
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

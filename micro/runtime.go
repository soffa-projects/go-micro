package micro

import (
	"embed"
	"github.com/fabriqs/go-micro/di"
	"github.com/fabriqs/go-micro/util/h"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Cfg struct {
	Name          string
	Version       string
	Features      []Feature
	FS            embed.FS
	DefaultLocale string
	Locales       string
	MultiTenant   bool
}

// ----------------------------------------------

func Set(key string, value string) {
	os.Setenv(key, value)
}

func (app *App) Get(key string) string {
	return h.RequireEnv(key)
}

func (app *App) Cleanup() {
	if app.Env.DB != nil {
		app.Env.DB.Close()
	}
}

func (app *App) Init(features []Feature) *App {
	//env.components = make([]Component, 0)

	env := app.Env
	globalLocalizer = env.Localizer

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

	for _, feat := range features {
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

func (app *App) Run(addr ...string) {
	// setup exit code for graceful shutdown
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	var port string
	if len(addr) == 0 {
		port = h.GetEnvOrDefault("PORT", "8080")
	} else {
		port = addr[0]
	}

	// start the server
	go func() {
		_ = app.Env.Router.Start("0.0.0.0:" + port)
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = app.Env.Router.Shutdown()
		if app.Env.DB != nil {
			app.Env.DB.Close()
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

func gracefully() {
	quit := make(chan os.Signal, 1)
	defer close(quit)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func T(messageId string, other ...string) string {
	if globalLocalizer == nil {
		return messageId
	}
	theOrder := ""
	if len(other) > 0 {
		theOrder = other[0]
	}
	return globalLocalizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    messageId,
			Other: theOrder,
		},
	})
}

package micro

import (
	"context"
	"embed"
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/di"
	"github.com/soffa-projects/go-micro/util/h"
	"github.com/swaggo/swag"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Cfg struct {
	//Name            string
	//Version         string
	Features                   []Feature
	FS                         embed.FS
	DefaultLocale              string
	AvailableLocales           []string
	MultiTenant                bool
	TablePrefix                string
	EnableDiscovery            bool
	BasePath                   string
	DisableRouter              bool
	DisableJwtFilter           bool
	CorsDisabled               bool
	DisableImplicitTransaction bool
	SwaggerSpec                *swag.Spec
}

// ----------------------------------------------

var _configStore = map[string]string{}

func Set(key string, value string) {
	_configStore[key] = value
}

func Get(key string) string {
	return _configStore[key]
}

func (app *App) Cleanup() {
	if app.Env.DataSources != nil {
		app.Env.Close()
	}
}

func (app *App) Init(features []Feature) *App {
	//env.components = make([]Component, 0)

	env := app.Env
	globalLocalizer = env.Localizer

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
		_ = Subscribe(NotificationTopic, func(ctx Ctx, payload Event) error {
			return env.Notifier.Send(ctx, Notification{
				Message: payload.Event,
			})
		})
	}

	for _, feat := range features {
		if feat.Configure != nil {
			if err := feat.Configure(app); err != nil {
				log.Fatalf("failed to init feature %s.\n%v", feat.Name, err)
			}
		} else if feat.Init != nil {
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

func (app *App) AddShutdownListener(listener func()) {
	app.ShutdownListeners = append(app.ShutdownListeners, listener)
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
		_ = app.Router.Start("0.0.0.0:" + port)
	}()

	// run the cleanup after the server is terminated
	defer func() {
		_ = app.Router.Shutdown()
		if app.Env.DataSources != nil {
			app.Env.Close()
		}
		if app.ShutdownListeners != nil {
			for _, listener := range app.ShutdownListeners {
				listener()
			}
		}
	}()

	if app.Env.Scheduler != nil && !app.Env.Scheduler.IsEmpty() {
		go func() {
			time.Sleep(10 * time.Second)
			app.Env.Scheduler.StartAsync()
			log.Infof("scheduler started")
		}()
	}

	if app.Env.RedisClient != nil && app.Env.DiscoverySericeName != "" {
		go func() {
			rc := app.Env.RedisClient
			h.RaiseAny(rc.Set(context.Background(), app.Env.DiscoverySericeName, app.Env.DiscoveryServiceUrl, 0).Err())
			rc.Publish(context.Background(), DiscoveryServicesChannel, fmt.Sprintf(
				"%s:%s",
				app.Name,
				app.Env.DiscoveryServiceUrl,
			))
			log.Infof("discovery service url broadcasted: %s -> %s", app.Name, app.Env.DiscoveryServiceUrl)
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
	theOrder := messageId
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

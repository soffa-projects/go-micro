package adapters

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/util/h"
	"golang.org/x/text/language"
	"os"
	"strings"
)

func NewApp(name string, version string, cfg micro.Cfg) *micro.App {
	isProduction := os.Getenv("GO_ENV") == "production"
	if !isProduction {
		err := godotenv.Load()
		if err != nil {
			log.Warn("unable to loading .env file")
		}
	}

	env := &micro.Env{
		Production: isProduction,
	}
	setupLocales(env, cfg)
	prepareMultiTenancy(env, cfg)
	setupDatabase(env, cfg)
	setupScheduler(env)
	setupMailer(env)
	setupNotifications(env)
	setupTokenProvider(env)
	setupRouter(env, cfg)

	// configure locales if any
	return &micro.App{
		Name:    name,
		Version: version,
		Env:     env,
	}

}

func setupLocales(env *micro.Env, cfg micro.Cfg) {
	exists, localesFS := h.CheckFsFolder(cfg.FS, "locales")
	if !exists {
		log.Info("no config/locales found, skipping")
		return
	}
	bundle := i18n.NewBundle(language.French)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	locales := strings.Split(cfg.Locales, ",")
	for _, lang := range locales {
		_, err := bundle.LoadMessageFileFS(localesFS, fmt.Sprintf("locale.%s.toml", lang))
		if err != nil {
			panic(err)
		}
	}
	localizer := i18n.NewLocalizer(bundle, locales...)
	log.Infof("%d locales loaded", len(cfg.Locales))
	env.Localizer = localizer

}

func prepareMultiTenancy(env *micro.Env, cfg micro.Cfg) {
	var tenantLoader micro.TenantLoader
	if cfg.MultiTenant {
		defaultTenants := h.RequireEnv(micro.DatabaseInitialTenants)
		tenantLoader = micro.NewFixedTenantLoader(strings.Split(defaultTenants, ","))
	} else {
		tenantLoader = micro.NewFixedTenantLoader([]string{micro.DefaultTenantId})
	}
	env.TenantLoader = tenantLoader
}

func setupDatabase(env *micro.Env, cfg micro.Cfg) {
	exists, migrationsFS := h.CheckFsFolder(cfg.FS, "db/migrations")
	if !exists {
		log.Info("no config/db/migrations found, skipping")
		return
	}
	databaseUrl := h.RequireEnv(micro.DatabaseUrl)

	if env.TenantLoader == nil {
		env.TenantLoader = micro.NewFixedTenantLoader([]string{micro.DefaultTenantId})
	}

	tenants := env.TenantLoader.GetTenant()
	links := map[string]micro.DataSource{}

	if cfg.MultiTenant {
		for _, tenant := range tenants {
			links[tenant] = NewGormAdapter(databaseUrl, tenant)
			if tenant == micro.DefaultTenantId {
				links[tenant].Migrate(migrationsFS, "shared")
			} else {
				links[tenant].Migrate(migrationsFS, "tenant")
			}
		}
	} else {
		for _, tenant := range tenants {
			links[tenant] = NewGormAdapter(databaseUrl, tenant)
			links[tenant].Migrate(migrationsFS, ".")
		}
	}

	env.DB = links
}

func setupScheduler(env *micro.Env) {
	env.Scheduler = NewGoCronAdapter(env.TenantLoader)
}

func setupMailer(env *micro.Env) {
	config := h.GetEnv(micro.EmailSender, "MAILER")
	if config == "" {
		log.Infof("env.%s is empty, skipping mailer setup", micro.EmailSender)
		return
	}
	var mailer micro.Mailer
	mailerConfig := strings.Split(config, "://")
	if mailerConfig[0] == "sendgrid" {
		mailer = NewSendGridEmailSender(mailerConfig[1])
	} else if mailerConfig[0] == "fake" {
		mailer = NewFakeEmailSender()
	} else {
		log.Fatalf("mailer provider not supported: %s", mailerConfig)
	}
	env.Mailer = mailer
}

func setupNotifications(env *micro.Env) {

	config := h.GetEnv(micro.NotificationSender, "NOTIFIER")
	if config == "" {
		log.Info("env.NOTIFICATION_SENDER is empty, skipping notifications setup")
		return
	}

	var service micro.NotificationService

	if strings.Contains(config, "discord.com") {
		service = NewDiscordClient(config)
	} else if strings.Contains(config, "noop") {
		service = micro.NewNoopNotificationService()
	} else {
		log.Fatalf("notifications manager provider not supported: %s", config)
	}

	env.Notifier = service

}

func setupTokenProvider(env *micro.Env) {
	secret := h.GetEnv(micro.ServerToken)
	if secret == "" {
		log.Infof("env.%s is empty, skipping token provider setup", micro.ServerToken)
		return
	}
	env.TokenProvider = micro.NewTokenProvider(secret)
}

func setupRouter(env *micro.Env, cfg micro.Cfg) {
	router := NewEchoAdapter(
		micro.RouterConfig{
			Cors:             true,
			SentryDsn:        h.GetEnv("SENTRY_DSN"),
			RemoveTrailSlash: true,
			BodyLimit:        "2M",
			Swagger:          true,
			TokenProvider:    env.TokenProvider,
			MultiTenant:      cfg.MultiTenant,
		})
	env.Router = router
}

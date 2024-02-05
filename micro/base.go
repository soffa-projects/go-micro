package micro

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/di"
)

var DefaultTenantId = "public"
var TenantIdHttpHeader = "X-TenantId"

type Feature struct {
	Name string
	// Deprecated: Use Configure instead.
	Init      func(app *App) (di.Component, error)
	Configure func(app *App) error
}

type App struct {
	Name              string
	Version           string
	Env               *Env
	ShutdownListeners []func()
}

type AuthToken struct {
	Issuer   string `json:"token"`
	Audience string `json:"audience"`
}

type Authentication struct {
	Token         *AuthToken
	Authenticated bool
	Name          string
	Email         string
	UserId        string
	PhonerNumber  string
	Claims        map[string]interface{}
	TenantId      string
	Roles         []string
	Permissions   []string
	IpAddress     string
}

func (a *Authentication) Claim(key string) interface{} {
	if a.Claims == nil {
		return nil
	}
	if v, ok := a.Claims[key]; ok {
		return v
	}
	return nil
}

type TenantLoader interface {
	GetTenant() []string
}

type ProxyCtx struct {
	Ctx
	UpstreamId    string
	UpstreamUrl   string
	Authorization string
	Bearer        string
}

type Ctx struct {
	TenantId string
	Auth     *Authentication
	db       DataSource
}

type Env struct {
	Ctx
	Conf                interface{}
	AppName             string
	AppVersion          string
	DB                  map[string]DataSource
	ServerPort          int
	Router              Router
	Scheduler           Scheduler
	TokenProvider       TokenProvider
	Notifier            NotificationService
	Mailer              Mailer
	Production          bool
	TenantLoader        TenantLoader
	Localizer           *i18n.Localizer
	RedisClient         *redis.Client
	DiscoverySericeName string
	DiscoveryServiceUrl string
}

type AppCfg struct {
	Name     string
	Features []Feature
	Router   Router
	DB       DataSource
}

type FixedTenantLoader struct {
	TenantLoader
	tenants []string
}

func (f *FixedTenantLoader) GetTenant() []string {
	return f.tenants
}

func NewFixedTenantLoader(tenants []string) *FixedTenantLoader {
	return &FixedTenantLoader{tenants: tenants}
}

func (ctx Ctx) IsDefaultTenant() bool {
	return ctx.TenantId == DefaultTenantId
}

func (ctx Ctx) IsAuthenticated() bool {
	if ctx.Auth == nil {
		return false
	}
	return ctx.Auth.Authenticated
}

func NewCtx(tenantId string) Ctx {
	var db DataSource
	if db == nil && globalEnv != nil && globalEnv.DB != nil {
		db = globalEnv.DB[tenantId]
	}
	return Ctx{
		TenantId: tenantId,
		db:       db,
	}
}

func (e Env) DefaultDB() DataSource {
	db := e.DB[DefaultTenantId]
	return db
}

func NewAuthCtx(auth *Authentication) Ctx {
	if auth == nil {
		return NewCtx(DefaultTenantId)
	}
	ctx := NewCtx(auth.TenantId)
	ctx.Auth = auth
	return ctx
}

func (ctx Ctx) Tx(cb func(tx Ctx) error) error {
	db := ctx.db
	if db == nil {
		db = globalEnv.DB[ctx.TenantId]
	}
	if db == nil {
		log.Warn("no db found in current context (skipping global transaction)")
		return cb(Ctx{
			TenantId: ctx.TenantId,
			Auth:     ctx.Auth,
		})
	}
	return db.Transaction(func(tx DataSource) error {
		return cb(Ctx{
			TenantId: ctx.TenantId,
			Auth:     ctx.Auth,
			db:       tx,
		})
	})
}

func (e Env) Close() {
	for _, db := range e.DB {
		db.Close()
	}
}

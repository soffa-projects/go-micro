package micro

import (
	"github.com/fabriqs/go-micro/di"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var DefaultTenantId = "public"

type Feature struct {
	Name string
	Init func(app *App) (di.Component, error)
}

type App struct {
	Name    string
	Version string
	Env     *Env
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
	TenantId      string
	Roles         []string
	Permissions   []string
}

type TenantLoader interface {
	GetTenant() []string
}

type Ctx struct {
	TenantId string
	Auth     *Authentication
	db       DataSource
}

type Env struct {
	Ctx
	Conf          interface{}
	DB            map[string]DataSource
	Router        Router
	Scheduler     Scheduler
	TokenProvider TokenProvider
	Notifier      NotificationService
	Mailer        Mailer
	Production    bool
	TenantLoader  TenantLoader
	Localizer     *i18n.Localizer
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

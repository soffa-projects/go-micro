package micro

import (
	"github.com/fabriqs/go-micro/di"
)

var DefaltTenantId = "public"

type Feature struct {
	Name string
	Init func(app *App) (di.Component, error)
}

type App struct {
	Name     string
	Version  string
	Config   interface{}
	Features []Feature
	Env      *Env
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
}

type Env struct {
	Ctx
	Conf          interface{}
	DataSource    DataSource
	Router        Router
	Scheduler     Scheduler
	TokenProvider TokenProvider
	Notifier      NotificationService
	Mailer        Mailer
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
	return Ctx{
		TenantId: tenantId,
	}
}

package micro

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/di"
	"github.com/soffa-projects/go-micro/schema"
	"github.com/soffa-projects/go-micro/util/h"
	"net/http"
	"reflect"
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
	Router            Router
}

type AuthToken struct {
	Token    string `json:"token"`
	Issuer   string `json:"issuer"`
	Audience string `json:"audience"`
}

type Authentication struct {
	Authenticated bool
	Authorization string
	Bearer        string
	Username      string
	Password      string
	Name          string
	Email         string
	UserId        string
	PhonerNumber  string
	Claims        map[string]interface{}
	//TenantId      string
	Roles       []string
	Permissions []string
	IpAddress   string
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
	Env      *Env
	db       DataSource
	Wrapped  interface{}
}

type Env struct {
	Ctx
	Conf        interface{}
	AppName     string
	AppVersion  string
	DataSources map[string]DataSource
	ServerPort  int
	//
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

func (ctx Ctx) TenantDB(tenant string) (DataSource, error) {
	if ds, ok := ctx.Env.DataSources[tenant]; !ok {
		return nil, errors.New("missing_db_tenant")
	} else {
		return ds, nil
	}
}
func (e Env) TenantDB(tenant string) (DataSource, error) {
	if ds, ok := e.DataSources[tenant]; !ok {
		return nil, errors.New("missing_db_tenant")
	} else {
		return ds, nil
	}
}

func (ctx Ctx) CurrentDB() DataSource {
	return ctx.db
}

func (ctx Ctx) IsAuthenticated() bool {
	if ctx.Auth == nil {
		return false
	}
	return ctx.Auth.Authenticated
}

func (ctx Ctx) Authenticate(auth Authentication, tenant string) error {
	ctx.Auth = &Authentication{}
	if err := h.CopyAllFields(ctx.Auth, auth, true); err != nil {
		return err
	}
	e := ctx.Wrapped.(echo.Context)
	if tenant != "" {
		e.Set(TenantId, tenant)
	}
	e.Set(AuthKey, ctx.Auth)
	return nil
}

func (e Env) SharedDB() DataSource {
	ds := e.DataSources[DefaultTenantId]
	if ds == nil {
		log.Fatalf("no shared db found")
	}
	return ds
}

func NewCtx(env *Env, tenantId string) Ctx {
	var db DataSource
	if db == nil && env != nil && env.DataSources != nil {
		db = env.DataSources[tenantId]
	}
	return Ctx{
		TenantId: tenantId,
		db:       db,
		Env:      env,
	}
}

/*func (e Env) DefaultDB() DataSource {
	if db, ok := e.DB[DefaultTenantId]; ok {
		return db
	}
	log.Fatalf("no default db found, missing tenant :%s", DefaultTenantId)
	return nil
}*/

func NewAuthCtx(env *Env, tenantId string, auth *Authentication) Ctx {
	if auth == nil {
		return NewCtx(env, DefaultTenantId)
	}
	ctx := NewCtx(env, tenantId)
	ctx.Auth = auth
	return ctx
}

func (ctx Ctx) Tx(cb func(tx Ctx) error) error {
	db := ctx.db
	if db == nil {
		db = ctx.Env.DataSources[ctx.TenantId]
	}
	if db == nil {
		log.Warn("no db found in current context (skipping global transaction)")
		return cb(Ctx{
			TenantId: ctx.TenantId,
			Auth:     ctx.Auth,
			Env:      ctx.Env,
			Wrapped:  ctx.Wrapped,
		})
	}
	return db.Transaction(func(tx DataSource) error {
		return cb(Ctx{
			TenantId: ctx.TenantId,
			Auth:     ctx.Auth,
			db:       tx,
			Env:      ctx.Env,
			Wrapped:  ctx.Wrapped,
		})
	})
}

func (e Env) Close() {
	for _, db := range e.DataSources {
		db.Close()
	}
}

func (ctx Ctx) Response() *echo.Response {
	e := ctx.Wrapped
	if e == nil {
		return nil
	}
	return e.(echo.Context).Response()
}

func (ctx Ctx) RequestWithValue(key string, value any) *http.Request {
	req := ctx.Request()
	return req.WithContext(context.WithValue(req.Context(), key, value))
}

func (ctx Ctx) Request() *http.Request {
	e := ctx.Wrapped
	if e == nil {
		return nil
	}
	return e.(echo.Context).Request()
}
func (ctx Ctx) SetTenantId(value string) {
	e := ctx.Wrapped
	if e == nil {
		return
	}
	e.(echo.Context).Set(TenantId, value)
	ctx.TenantId = value
}

func (ctx Ctx) Bind(modelType reflect.Type) (error, any) {
	if ctx.Wrapped == nil {
		log.Fatalf("ctx.Wrapped is nil")
		return nil, nil
	}
	entity := reflect.New(modelType).Interface()
	binder := &echo.DefaultBinder{}
	echoContext := ctx.Wrapped.(echo.Context)
	if echoContext == nil {
		log.Fatalf("ctx.Wrapped is not an echo.Context")
		return nil, nil
	}
	if err := binder.Bind(entity, echoContext); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			schema.ErrorResponse{
				Kind:    "input.binding",
				Message: "invalid_binding",
				Errors:  err.Error(),
			}), nil
	}
	return nil, entity
}

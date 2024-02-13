package micro

import (
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/swag"
	"net/http"
	"strings"
)

// AuthKey is used in adapters
//
//goland:noinspection GoUnusedConst
const AuthKey = "user"
const TenantId = "tenant"
const EnvKey = "env"
const DisableImplicitTransaction = "implicit_transaction_disabled"

type Router interface {
	BaseRouter
	Handler() http.Handler
	Start(addr string) error
	Shutdown() error
	Group(path string, filters ...MiddlewareFunc) BaseRouter
	Use(filter MiddlewareFunc)
	Proxy(path string, upstreams *RouterUpstream, filters ...MiddlewareFunc)
}

type BaseRouter interface {
	POST(path string, handler any, filters ...MiddlewareFunc)
	PUT(path string, handler any, filters ...MiddlewareFunc)
	PATCH(path string, handler any, filters ...MiddlewareFunc)
	GET(path string, handler any, filters ...MiddlewareFunc)
	DELETE(path string, handler any, filters ...MiddlewareFunc)
	Any(path string, handler any, filters ...MiddlewareFunc)
	//Resource(resource string, model interface{})
}

type JwtCfg struct {
	Provider TokenProvider
}

type RouterConfig struct {
	Cors                       bool
	RemoveTrailSlash           bool
	BasePath                   string
	DisableImplicitTransaction bool
	BodyLimit                  string
	SwaggerSpec                *swag.Spec
	MultiTenant                bool
	//Prometheus       *PrometheusCfg
	//JwtAuth    bool
	Production       bool
	TokenProvider    TokenProvider
	DisableJwtFilter bool
	SentryDsn        string
	OnShutdown       func()
}

type MiddlewareFunc func(ctx Ctx) error

type RouteFilter func(handler HandlerFunc) HandlerFunc

type HandlerFunc func(c Ctx) (any, error)

type ProxyHandlerFunc func(c *ProxyCtx) error

type UpstreamCtx struct {
	Authorization string
}

type ErrorResponse struct {
	Kind    string `json:"kind,omitempty"`
	Error   string `json:"error,omitempty"`
	Details any    `json:"details,omitempty"`
}

type RouterUpstream struct {
	data map[string]*Upstream
}

type Upstream struct {
	Id     string
	Uri    string
	Prefix string
	Strip  bool
}

func NewRouterUpstream(data map[string]*Upstream) *RouterUpstream {
	for id, up := range data {
		up.Id = id
	}
	return &RouterUpstream{
		data: data,
	}
}

func (u *RouterUpstream) SetUri(id string, value string) {
	u.data[id].Uri = value
	log.Infof("upstream updated: %s --> %s", id, value)
}

func (u *RouterUpstream) Lookup(path string) *Upstream {
	for _, up := range u.data {
		if strings.HasPrefix(path, up.Prefix) {
			return up
		}
	}
	return nil
}

func (u *RouterUpstream) All() map[string]*Upstream {
	return u.data
}

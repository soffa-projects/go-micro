package micro

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// AuthKey is used in adapters
//
//goland:noinspection GoUnusedConst
const AuthKey = "user"
const TenantId = "tenant"

type Router interface {
	BaseRouter
	Handler() http.Handler
	Start(addr string) error
	Shutdown() error
	Group(path string, filters ...MiddlewareFunc) BaseRouter
	Proxy(path string, upstreams *RouterUpstream, handler ProxyHandlerFunc)
}

type BaseRouter interface {
	POST(path string, handler interface{}, filters ...MiddlewareFunc)
	PUT(path string, handler interface{}, filters ...MiddlewareFunc)
	PATCH(path string, handler interface{}, filters ...MiddlewareFunc)
	GET(path string, handler interface{}, filters ...MiddlewareFunc)
	DELETE(path string, handler interface{}, filters ...MiddlewareFunc)
	Any(path string, handler interface{}, filters ...MiddlewareFunc)
	//Resource(resource string, model interface{})
}

type JwtCfg struct {
	Provider TokenProvider
}

type RouterConfig struct {
	Cors             bool
	RemoveTrailSlash bool
	BodyLimit        string
	Swagger          bool
	MultiTenant      bool
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

type ProxyHandlerFunc func(c ProxyCtx) (*UpstreamCtx, error)

type UpstreamCtx struct {
	Authorization string
}

type ErrorResponse struct {
	Kind    string `json:"kind,omitempty"`
	Error   string `json:"error,omitempty"`
	Details any    `json:"details,omitempty"`
}

type RouterUpstream struct {
	data map[string]string
}

func NewRouterUpstream(data map[string]string) *RouterUpstream {
	return &RouterUpstream{
		data: data,
	}
}

func (u *RouterUpstream) Set(key string, value string) {
	u.data[key] = value
	log.Infof("upstream updated: %s --> %s", key, value)
}

func (u *RouterUpstream) Lookup(path string) string {
	for p, up := range u.data {
		if strings.HasPrefix(path, p) {
			log.Infof("upstream found: %s --> %s", path, up)
			return fmt.Sprintf("%s%s", up, strings.TrimPrefix(path, p))
		}
	}
	return ""
}

func (u *RouterUpstream) All() map[string]string {
	return u.data
}

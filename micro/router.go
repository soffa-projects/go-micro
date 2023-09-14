package micro

import (
	"net/http"
)

// AuthKey is used in adapters
//
//goland:noinspection GoUnusedConst
const AuthKey = "user"

type Router interface {
	BaseRouter
	Handler() http.Handler
	Start(addr string) error
	Shutdown() error
	Group(path string, filters ...MiddlewareFunc) BaseRouter
}

type BaseRouter interface {
	POST(path string, handler interface{}, filters ...MiddlewareFunc)
	PUT(path string, handler interface{}, filters ...MiddlewareFunc)
	PATCH(path string, handler interface{}, filters ...MiddlewareFunc)
	GET(path string, handler interface{}, filters ...MiddlewareFunc)
	DELETE(path string, handler interface{}, filters ...MiddlewareFunc)
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
	TokenProvider TokenProvider
	SentryDsn     string
	OnShutdown    func()
}

type MiddlewareFunc func(ctx Ctx) error

type RouteFilter func(handler HandlerFunc) HandlerFunc

type HandlerFunc func(c Ctx) (any, error)

type ErrorResponse struct {
	Kind    string `json:"kind,omitempty"`
	Error   string `json:"error,omitempty"`
	Details any    `json:"details,omitempty"`
}

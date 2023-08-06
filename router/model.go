package router

import (
	"github.com/fabriqs/go-micro/crypto"
	"github.com/fabriqs/go-micro/schema"
	"net/http"
)

// AuthKey is used in adapters
//
//goland:noinspection GoUnusedConst
const AuthKey = "user"

type Ctx interface {
	Bind(interface{}) error
	Send(interface{}) error
	NewTechnicalError(string) error
	Forbidden(message string) error
	Unauthorized(message string) error
	GetAuthentication() *schema.Authentication
	IsAuthenticated() bool
	// RequireAuthentication()
}

type R interface {
	Base
	Handler() http.Handler
	Start(addr string) error
	Shutdown() error
	Group(path string, filters ...MiddlewareFunc) Base
}

type Base interface {
	POST(path string, handler HandlerFunc, filters ...MiddlewareFunc)
	PUT(path string, handler HandlerFunc, filters ...MiddlewareFunc)
	PATCH(path string, handler HandlerFunc, filters ...MiddlewareFunc)
	GET(path string, handler HandlerFunc, filters ...MiddlewareFunc)
	DELETE(path string, handler HandlerFunc, filters ...MiddlewareFunc)
}

type JwtCfg struct {
	Provider crypto.JwtProvider
}

type Cfg struct {
	Production       bool
	Cors             bool
	RemoveTrailSlash bool
	BodyLimit        string
	Swagger          bool
	//Prometheus       *PrometheusCfg
	JwtAuth    bool
	SentryDsn  string
	OnShutdown func()
}

type MiddlewareFunc func(ctx Ctx) error

type RouteFilter func(handler HandlerFunc) HandlerFunc

type HandlerFunc func(Ctx) (any, error)

type ErrorResponse struct {
	Kind    string `json:"kind,omitempty"`
	Error   string `json:"error,omitempty"`
	Details any    `json:"details,omitempty"`
}

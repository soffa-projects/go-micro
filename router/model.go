package router

import (
	"github.com/fabriqs/go-micro/crypto"
	"github.com/fabriqs/go-micro/h"
	"net/http"
)

const AuthKey = "user"

type Ctx interface {
	Bind(interface{}) error
	Send(interface{}) error
	// TODO() error
	NewTechnicalError(string) error
	Forbidden(message string) error
	Unauthorized(message string) error
	GetAuthentication() *Authentication
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

type AuthToken struct {
	Issuer   string `json:"token"`
	Audience string `json:"audience"`
}

type Authentication struct {
	Token         *AuthToken
	Authenticated bool
	Username      string
	Email         string
	UserId        string
	Roles         []string
	Permissions   []string
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

func TODO() interface{} {
	return h.Map{
		"status": "TODO",
	}
}

type MiddlewareFunc func(ctx Ctx) error

type RouteFilter func(handler HandlerFunc) HandlerFunc

type HandlerFunc func(Ctx) (interface{}, error)

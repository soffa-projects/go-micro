package adapters

import (
	"context"
	"fmt"
	"github.com/getsentry/sentry-go"
	sentryecho "github.com/getsentry/sentry-go/echo"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/schema"
	"github.com/soffa-projects/go-micro/util/digest"
	"github.com/soffa-projects/go-micro/util/errors"
	"github.com/soffa-projects/go-micro/util/h"
	echoSwagger "github.com/swaggo/echo-swagger"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

var validate *validator.Validate

//goland:noinspection GoTypeAssertionOnErrors
func Bind(c echo.Context, input interface{}) error {

	binder := &echo.DefaultBinder{}
	if err := binder.Bind(input, c); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			schema.ErrorResponse{
				Kind:    "input.bindind",
				Message: "invalid_request_payload",
				Errors:  err.Error(),
			})
	}
	if err := binder.BindHeaders(c, input); err != nil {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			schema.ErrorResponse{
				Kind:    "input.bindind",
				Message: "invalid_request_payload",
				Errors:  err.Error(),
			})
	}

	if err := validate.Struct(input); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errs := map[string]string{}
		for _, e := range validationErrors {
			errs[e.Field()] = e.Tag()
		}
		return echo.NewHTTPError(
			http.StatusBadRequest,
			schema.ErrorResponse{
				Kind:    "validation",
				Message: "validation.failed",
				Errors:  errs,
			},
		)
	}

	return nil
}

// =================================================================================
// ECHO ROUTER ADAPTER
// =================================================================================

type echoRouterAdapter struct {
	micro.Router
	e *echo.Echo
}

func NewEchoAdapter(env *micro.Env, config micro.RouterConfig) micro.Router {
	e := echo.New()
	e.HideBanner = true
	e.IPExtractor = echo.ExtractIPFromXFFHeader()

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(micro.EnvKey, env)
			if config.DisableImplicitTransaction {
				c.Set(micro.DisableImplicitTransaction, "1")
			}
			tenantId := c.Request().Header.Get(micro.TenantIdHttpHeader)
			if tenantId == "" {
				tenantId = micro.DefaultTenantId
			}
			ipAddress := c.RealIP()
			if h.IsStrEmpty(ipAddress) {
				ipAddress = c.Request().RemoteAddr
			}
			if h.IsStrEmpty(ipAddress) {
				ipAddress = "0.0.0.0"
			}
			authzHeader := c.Request().Header.Get("Authorization")
			bearer := ""
			if authzHeader != "" && strings.HasPrefix(strings.ToLower(authzHeader), "bearer ") {
				bearer = authzHeader[7:]
			}
			auth := &micro.Authentication{
				Authenticated: false,
				Authorization: authzHeader,
				Bearer:        bearer,
				IpAddress:     ipAddress,
			}
			if authzHeader != "" && strings.HasPrefix(strings.ToLower(authzHeader), "basic ") {
				parts := strings.SplitN(digest.DecodeBase64(authzHeader[6:]), ":", 2)
				if len(parts) >= 2 {
					auth.Username = parts[0]
					auth.Password = parts[1]
				}
			}
			c.Set(micro.TenantId, tenantId)
			c.Set(micro.AuthKey, auth)
			return next(c)
		}
	})
	/*e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
	}))*/
	e.Use(middleware.RequestID())
	if config.Cors {
		e.Use(middleware.CORS())
	}
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool {
			return strings.Contains(c.Request().URL.Path, ".")
		},
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	if config.Production {
		e.Use(middleware.Recover())
	}
	if config.SentryDsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn: config.SentryDsn,
			// Set TracesSampleRate to 1.0 to capture 100%
			// of transactions for performance monitoring.
			// We recommend adjusting this value in production,
			// TracesSampleRate: 1.0,
		}); err != nil {
			fmt.Printf("Sentry initialization failed: %v\n", err)
		}
		e.Use(sentryecho.New(sentryecho.Options{}))
	}
	if config.BodyLimit != "" {
		e.Use(middleware.BodyLimit(config.BodyLimit))
	}
	if config.Swagger {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
	/*if config.Prometheus != nil && config.Prometheus.Enabled {
		e.F(echoprometheus.NewMiddleware(config.Prometheus.Subsystem)) // adds middleware to gather metrics
		e.GET("/metrics", echoprometheus.NewHandler()) // adds route to serve gathered metrics
	}*/
	if config.RemoveTrailSlash {
		e.Pre(middleware.RemoveTrailingSlash())
	}

	if config.TokenProvider != nil && !config.DisableJwtFilter {
		e.Use(echojwt.WithConfig(echojwt.Config{
			SigningKey: []byte(config.TokenProvider.SigningKey()),
			ContextKey: micro.AuthKey,
			Skipper: func(c echo.Context) bool {
				// let the app decide which routes require authentication
				authz := c.Request().Header.Get("Authorization")
				if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
					return true
				}
				parts := strings.Split(authz, ".")
				if len(parts) == 3 {
					// It might be a JWT
					return false
				}
				// skip also if the token is not a jwt
				return true
			},
			ErrorHandler: func(c echo.Context, err error) error {
				return c.JSON(http.StatusUnauthorized, err.Error())
			},
			SuccessHandler: func(c echo.Context) {
				user := c.Get(micro.AuthKey)
				tenantId := c.Get(micro.TenantId)
				if user == nil {
					return
				}
				token := user.(*jwt.Token)

				claims0 := token.Claims

				sub, _ := claims0.GetSubject()
				// issuer, _ := claims0.GetIssuer()

				//tenantId := micro.DefaultTenantId
				/*if config.MultiTenant && issuer != "" {
					tenantId = issuer
				}*/

				ipAddress := c.RealIP()
				if h.IsStrEmpty(ipAddress) {
					ipAddress = c.Request().RemoteAddr
				}
				if h.IsStrEmpty(ipAddress) {
					ipAddress = "0.0.0.0"
				}
				auth := &micro.Authentication{
					UserId:        sub,
					Authenticated: true,
					IpAddress:     ipAddress,
				}

				//TODO: F depenedency injection (UserService)
				if claims0 != nil && reflect.TypeOf(claims0).String() == "jwt.MapClaims" {
					claims := claims0.(jwt.MapClaims)

					if value, ok := h.MapLookup(claims, "tenant", "tenant_id", "tenantId"); ok {
						tenantId = value.(string)
					}
					if value, ok := h.MapLookup(claims, "roles", "role"); ok {
						auth.Roles = strings.Split(value.(string), ",")
					}
					if value, ok := claims["permissions"]; ok {
						auth.Permissions = strings.Split(value.(string), ",")
					}
					if value, ok := claims["name"]; ok {
						auth.Name = value.(string)
					}
					if value, ok := claims["email"]; ok {
						auth.Email = value.(string)
					}
					if value, ok := h.MapLookup(claims, "phone", "phone_number", "phoneNumber"); ok {
						auth.PhonerNumber = value.(string)
					}
					auth.Claims = make(map[string]interface{})
					for key, value := range claims {
						if h.IsNotEmpty(value) && !h.IsStrEmpty(value) {
							auth.Claims[key] = value
						}
					}
				}

				c.Set(micro.TenantId, tenantId)
				c.Set(micro.AuthKey, auth)
			},
		}))
	}

	e.GET("/health", func(c echo.Context) error {
		status := schema.NewHealthStatus()
		return c.JSON(http.StatusOK, status)
	})

	return &echoRouterAdapter{e: e}
}

func (r *echoRouterAdapter) Handler() http.Handler {
	return r.e
}

func (r *echoRouterAdapter) Start(addr string) error {
	return r.e.Start(addr)
}

func (r *echoRouterAdapter) Shutdown() error {
	return r.e.Shutdown(context.Background())
}

func (r *echoRouterAdapter) GET(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodGet, path, handler, filters)
}

func (r *echoRouterAdapter) POST(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodPost, path, handler, filters)
}

func (r *echoRouterAdapter) PUT(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodPut, path, handler, filters)
}

func (r *echoRouterAdapter) PATCH(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodPatch, path, handler, filters)
}

func (r *echoRouterAdapter) DELETE(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodDelete, path, handler, filters)
}

func (r *echoRouterAdapter) request(method string, path string, handler interface{}, filters []micro.MiddlewareFunc) {
	r.e.Match([]string{method}, path, func(c echo.Context) (err error) {
		defer func() {
			if err0 := recover(); err0 != nil {
				err = mapHttpResponse(c, err0.(error))
			}
		}()
		return handleRequest(c, handler)
	}, createMiddlewares(filters)...)
}

func (r *echoRouterAdapter) Group(path string, filters ...micro.MiddlewareFunc) micro.BaseRouter {
	return &echoGroupRoute{
		g: r.e.Group(path, createMiddlewares(filters)...),
	}
}

func (r *echoRouterAdapter) Proxy(path string, upstreams *micro.RouterUpstream) {
	r.e.Any(path, func(c echo.Context) error {
		uriParts := strings.Split(c.Request().URL.Path, "?")
		requestUri := uriParts[0]
		requestQuery := ""
		if len(uriParts) > 1 {
			requestQuery = uriParts[1]
		}
		upstream := upstreams.Lookup(requestUri)
		if upstream == nil {
			return echo.NewHTTPError(http.StatusNotFound, "no_upstream_found")
		}

		chain := h.F(url.JoinPath(upstream.Uri, strings.TrimPrefix(requestUri, upstream.Prefix)))
		if requestQuery != "" {
			chain = strings.Join([]string{chain, requestQuery}, "?")
		}
		req, _ := http.NewRequest(
			c.Request().Method,
			chain,
			c.Request().Body,
		)
		copyHeader(c.Request().Header, req.Header)

		authz := c.Get(micro.AuthKey)
		if authz != nil {
			_authz := authz.(*micro.Authentication)
			if _authz.Authenticated && _authz.Authorization != "" {
				req.Header.Set("Authorization", _authz.Authorization)
			}
		}

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		//goland:noinspection ALL
		defer resp.Body.Close()
		copyHeader(resp.Header, c.Response().Header())
		c.Response().WriteHeader(resp.StatusCode)
		_, err = io.Copy(c.Response().Writer, resp.Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, "Failed to send upstream response")
		}

		return nil
	})
}

func (r *echoRouterAdapter) Use(filter micro.MiddlewareFunc) {
	r.e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if err := filter(createRouteContext(c)); err != nil {
				return mapHttpResponse(c, err)
			}
			return next(c)
		}
	})
}

func copyHeader(src, dst http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// =================================================================================
// ECHO GROUP ROUTE
// =================================================================================

type echoGroupRoute struct {
	micro.BaseRouter
	g   *echo.Group
	ctx micro.Ctx
}

func (r *echoGroupRoute) GET(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodGet, path, handler, filters)
}

func (r *echoGroupRoute) POST(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodPost, path, handler, filters)
}

func (r *echoGroupRoute) PUT(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodPut, path, handler, filters)
}

func (r *echoGroupRoute) PATCH(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodPatch, path, handler, filters)
}

func (r *echoGroupRoute) DELETE(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request(http.MethodDelete, path, handler, filters)
}

func (r *echoGroupRoute) Any(path string, handler interface{}, filters ...micro.MiddlewareFunc) {
	r.request("*", path, handler, filters)
}

/*
func (r *echoGroupRoute) Resource(resource string, model any) {

	r.request(http.MethodPatch, "/:id", func(c micro.Ctx) any {
		err, entity := c.Bind(modelType)
		if e
	}, nil)

	r.request(http.MethodDelete, "/:id", func(c micro.Ctx, input IdValue) any {
		db := c.DB()
		if _, err := db.Delete(model, micro.Query{
			W:    "id = ?",
			Args: []interface{}{input.Id},
		}); err != nil {
			return err
		}
		return input
	}, nil)
}
*/

func (r *echoGroupRoute) request(method string, path string, handler interface{}, filters []micro.MiddlewareFunc) {

	if method == "*" {
		r.g.Any(path, func(c echo.Context) (err error) {
			defer func() {
				if err0 := recover(); err0 != nil {
					log.Error(err0)
					err = mapHttpResponse(c, err0.(error))
				}
			}()
			return mapHttpResponse(c, handleRequest(c, handler))
		}, createMiddlewares(filters)...)
		return
	}

	r.g.Match([]string{method}, path, func(c echo.Context) (err error) {
		defer func() {
			if err0 := recover(); err0 != nil {
				log.Error(err0)
				err = mapHttpResponse(c, err0.(error))
			}
		}()
		return mapHttpResponse(c, handleRequest(c, handler))

	}, createMiddlewares(filters)...)
}

// =================================================================================
// GENERIC HANDLER
// =================================================================================

//goland:noinspection GoTypeAssertionOnErrors
func handleRequest(c echo.Context, handler interface{}) (err error) {
	handlerType := reflect.TypeOf(handler)

	if handlerType.Kind() != reflect.Func {
		return mapHttpResponse(c, fmt.Errorf("controller method is not a function"))
	}

	numIn := handlerType.NumIn()
	if numIn == 0 {
		return mapHttpResponse(c, fmt.Errorf("controller method must have at least one argument (micro.Ctx)"))
	}
	if numIn > 2 {
		return mapHttpResponse(c, fmt.Errorf("controller method must have at most two arguments (micro.Ctx, input binding)"))
	}

	firstArgType := handlerType.In(0)

	if firstArgType != reflect.TypeOf(micro.Ctx{}) {
		return fmt.Errorf("handler must be a function with the first argument of type micro.Ctx")
	}

	ctx := createRouteContext(c)
	tenantId := ctx.TenantId
	if tenantId == "" {
		tenantId = micro.DefaultTenantId
	} else {
		log.Debugf("current tenant_id is %s", tenantId)
	}

	disabledTx := c.Get(micro.DisableImplicitTransaction)
	if disabledTx == "1" {
		return mapHttpResponse(c, invokeHandler(c, ctx, handler, handlerType, numIn))
	}
	return ctx.Tx(func(tx micro.Ctx) error {
		return mapHttpResponse(c, invokeHandler(c, ctx, handler, handlerType, numIn))
	})
}

func invokeHandler(c echo.Context, tx micro.Ctx, handler interface{}, handlerType reflect.Type, numIn int) error {
	args := []reflect.Value{reflect.ValueOf(tx)}

	if numIn == 2 {
		inputType := handlerType.In(1)
		inputValue := reflect.New(inputType).Elem()
		modelInput := inputValue.Addr().Interface() //
		if err := Bind(c, modelInput); err != nil {
			log.Errorf("validation failed for %s\n%v", c.Request().RequestURI, err.Error())
			return err
		}
		args = append(args, inputValue)
	}

	handlerValue := reflect.ValueOf(handler)

	res := handlerValue.Call(args)

	if len(res) > 2 {
		return fmt.Errorf("invalid handler return type")
	}

	var result interface{}

	for _, r := range res {
		if err, ok := r.Interface().(error); ok {
			return err
		} else if result == nil {
			result = r.Interface()
		}
	}
	if result == nil {
		return nil
	}
	return c.JSON(http.StatusOK, result)
}

func mapHttpResponse(c echo.Context, err error) error {
	if err == nil {
		return nil
	}
	log.Errorf("error while handling request %s -- %v", c.Request().RequestURI, err.Error())

	switch e := err.(type) {
	case *errors.FunctionalError:
		return c.JSON(http.StatusBadRequest, micro.ErrorResponse{
			Kind:    e.Kind,
			Error:   e.Message,
			Details: e.Details,
		})
	case *errors.TechnicalError:
		return c.JSON(http.StatusInternalServerError, micro.ErrorResponse{
			Kind:    e.Kind,
			Error:   e.Message,
			Details: e.Details,
		})
	case *errors.ForbiddenError:
		return c.JSON(http.StatusForbidden, micro.ErrorResponse{
			Kind:    e.Kind,
			Error:   e.Message,
			Details: e.Details,
		})
	case *errors.UnauthorizedError:
		return c.JSON(http.StatusUnauthorized, micro.ErrorResponse{
			Kind:    e.Kind,
			Error:   e.Message,
			Details: e.Details,
		})
	case *errors.ResourceNotFoundError:
		return c.JSON(http.StatusNotFound, micro.ErrorResponse{
			Kind:    e.Kind,
			Error:   e.Message,
			Details: e.Details,
		})
	case *errors.ConflictError:
		return c.JSON(http.StatusConflict, micro.ErrorResponse{
			Kind:    e.Kind,
			Error:   e.Message,
			Details: e.Details,
		})
	case *echo.HTTPError:
		return c.JSON(e.Code, e.Message)

	case *pgconn.PgError:
		if e.Code == "23505" {
			return c.JSON(http.StatusConflict, micro.ErrorResponse{
				Kind:  "db_error",
				Error: e.Message,
			})
		}
		return c.JSON(http.StatusInternalServerError, micro.ErrorResponse{
			Kind:  "db_error",
			Error: e.Message,
		})
	default:
		return c.JSON(http.StatusInternalServerError, micro.ErrorResponse{
			Kind:  "unknown_error",
			Error: err.Error(),
		})
	}

}

// =================================================================================
// INIT
// =================================================================================

func createMiddlewares(filters []micro.MiddlewareFunc) []echo.MiddlewareFunc {
	middlewares := make([]echo.MiddlewareFunc, 0)
	for _, filter := range filters {
		middlewares = append(middlewares, func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				if err := filter(createRouteContext(c)); err != nil {
					return mapHttpResponse(c, err)
				}
				return next(c)
			}
		})
	}
	return middlewares
}

func createRouteContext(c echo.Context) micro.Ctx {
	env := c.Get(micro.EnvKey).(*micro.Env)
	value := c.Get(micro.AuthKey)
	tenantId := c.Get(micro.TenantId).(string)
	var result micro.Ctx
	if value == nil {
		result = micro.NewCtx(env, tenantId)
	} else {
		auth := value.(*micro.Authentication)
		result = micro.NewAuthCtx(env, tenantId, auth)
	}
	result.Wrapped = c
	return result
}

func init() {
	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

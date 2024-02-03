package middleware

import (
	"fmt"
	"github.com/soffa-projects/go-micro/micro"
	"github.com/soffa-projects/go-micro/util/errors"
	"github.com/soffa-projects/go-micro/util/h"
	"strings"
)

// Authenticated returns a middleware that checks if the user is authenticated.
func Authenticated() micro.MiddlewareFunc {
	return func(ctx micro.Ctx) error {
		if !ctx.IsAuthenticated() {
			return errors.Unauthorized("Unauthorized")
		}
		return nil
	}
}

// AuthenticatedWithRole returns a middleware that checks if the user is authenticated and has the given role.
func AuthenticatedWithRole(roles ...string) micro.MiddlewareFunc {
	return func(ctx micro.Ctx) error {
		if !ctx.IsAuthenticated() {
			return errors.Unauthorized("Unauthorized")
		}
		auth := ctx.Auth
		for _, authRole := range auth.Roles {
			for _, role := range roles {
				if authRole == role {
					return nil
				}
			}
		}
		return errors.Forbidden(fmt.Sprintf("missing_role: %s", strings.Join(roles, ",")))
	}
}

func TenantRequired() micro.MiddlewareFunc {
	return func(ctx micro.Ctx) error {
		if h.IsStrEmpty(ctx.TenantId) || ctx.IsDefaultTenant() {
			return errors.Forbidden("TENANT_REQUIRED")
		}
		return nil
	}
}

package middleware

import (
	"fmt"
	"github.com/fabriqs/go-micro/router"
	"strings"
)

// Authenticated returns a middleware that checks if the user is authenticated.
func Authenticated() router.MiddlewareFunc {
	return func(ctx router.Ctx) error {
		if !ctx.IsAuthenticated() {
			return ctx.Unauthorized("Unauthorized")
		}
		return nil
	}
}

// AuthenticatedWithRole returns a middleware that checks if the user is authenticated and has the given role.
func AuthenticatedWithRole(roles ...string) router.MiddlewareFunc {
	return func(ctx router.Ctx) error {
		if !ctx.IsAuthenticated() {
			return ctx.Unauthorized("Unauthorized")
		}
		auth := ctx.GetAuthentication()
		for _, authRole := range auth.Roles {
			for _, role := range roles {
				if authRole == role {
					return nil
				}
			}
		}
		return ctx.Forbidden(fmt.Sprintf("missing_role: %s", strings.Join(roles, ",")))
	}
}

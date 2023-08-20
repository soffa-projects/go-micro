package middleware

import (
	"fmt"
	"github.com/fabriqs/go-micro/micro"
	"github.com/fabriqs/go-micro/util/errors"
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
		auth := ctx.Auth()
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

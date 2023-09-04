package bdd

import (
	"context"
	"net/http"
)

type authKey struct{}

type Authentication interface {
	ApplyHTTP(ctx context.Context, r *http.Request)
}

// UseAuthentication set pre configured authentication.
//
// Example:
//
//	bdd.UseAuthentication(ctx, &bdd.DefaultBasicAuthentication{
//	    User: "holy",
//	    Pass: "hell",
//	})
func UseAuthentication(ctx context.Context, auth Authentication) context.Context {
	return context.WithValue(ctx, authKey{}, auth)
}

func MustGetAuthentication(ctx context.Context) Authentication {
	auth := ctx.Value(authKey{}).(Authentication)
	return auth
}

func GetAuthentication(ctx context.Context) (Authentication, bool) {
	auth, ok := ctx.Value(authKey{}).(Authentication)
	return auth, ok
}

type DefaultBasicAuthentication struct {
	User string
	Pass string
}

func (d *DefaultBasicAuthentication) ApplyHTTP(ctx context.Context, r *http.Request) {
	r.SetBasicAuth(d.User, d.Pass)
}

type NoAuthAuthentication struct{}

func (n *NoAuthAuthentication) ApplyHTTP(ctx context.Context, r *http.Request) {
}

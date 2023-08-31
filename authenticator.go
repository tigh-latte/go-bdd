package bdd

import (
	"context"
	"net/http"
)

var authKey = struct{}{}

type Authentication interface {
	ApplyHTTP(ctx context.Context, r *http.Request)
}

// WithAuthentication set pre configured authentication.
//
// Example:
//
//	bdd.WithAuthentication(ctx, &bdd.DefaultBasicAuthentication{
//	    User: "holy",
//	    Pass: "hell",
//	})
func WithAuthentication(ctx context.Context, auth Authentication) context.Context {
	return context.WithValue(ctx, authKey, auth)
}

func MustGetAuthentication(ctx context.Context) Authentication {
	auth := ctx.Value(authKey).(Authentication)
	return auth
}

type DefaultBasicAuthentication struct {
	User string
	Pass string
}

func (d *DefaultBasicAuthentication) ApplyHTTP(ctx context.Context, r *http.Request) {
	r.SetBasicAuth(d.User, d.Pass)
}

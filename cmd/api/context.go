package main

import (
	"context"
	"net/http"

	"github.com/thats-insane/awt-final/internal/data"
)

type ctxKey string

const userCtxKey = ctxKey("user")

func (a *appDependencies) ctxSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userCtxKey, user)
	return r.WithContext(ctx)
}

func (a *appDependencies) ctxGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userCtxKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}

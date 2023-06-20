package main

import (
	"context"
	"greenlight.badrchoubai.dev/internal/data"
	"net/http"
)

type contextKey string

const userContextKey = contextKey("user")

func (application *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (application *application) contextGetUser(r *http.Request, user *data.User) *http.Request {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}

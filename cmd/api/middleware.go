package main

import (
	"errors"
	"fmt"
	"golang.org/x/time/rate"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/validator"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Global Rate Limiter
//func (app *application) rateLimit(next http.Handler) http.Handler {
//	limiter := rate.NewLimiter(2, 4)
//
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		if !limiter.Allow() {
//			app.rateLimitExceededResponse(w, r)
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}

// IP Capturing Limiter
func (application *application) rateLimiter(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			application.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(
					rate.Limit(application.config.limiter.rps),
					application.config.limiter.burst,
				),
			}
		}
		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			application.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (application *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				application.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (application *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = application.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			application.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			application.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := application.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				application.invalidAuthenticationTokenResponse(w, r)
			default:
				application.serverErrorResponse(w, r, err)
			}
			return
		}

		r = application.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

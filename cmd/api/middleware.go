package main

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"greenlight.badrchoubai.dev/internal/data"
	"greenlight.badrchoubai.dev/internal/validator"
	"net/http"
	"strconv"
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
		if application.config.limiter.enabled {
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(application.config.limiter.rps), application.config.limiter.burst),
				}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				application.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

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

func (application *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := application.contextGetUser(r)

		if user.IsAnonymous() {
			application.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func (application *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := application.contextGetUser(r)

		if !user.Activated {
			application.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return application.requireAuthenticatedUser(fn)
}

func (application *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := application.contextGetUser(r)

		permissions, err := application.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			application.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Include(code) {
			application.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return application.requireActivatedUser(fn)
}

func (application *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range application.config.trustedOrigins {
				if origin == application.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	headerWritten bool
}

func newMetricsResponseWrite(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (mrw *metricsResponseWriter) Header() http.Header {
	return mrw.ResponseWriter.Header()
}

func (mrw *metricsResponseWriter) WriteHeader(statusCode int) {
	mrw.ResponseWriter.WriteHeader(statusCode)
	if !mrw.headerWritten {
		mrw.statusCode = statusCode
		mrw.headerWritten = true
	}
}

func (mrw *metricsResponseWriter) Write(b []byte) (int, error) {
	mrw.headerWritten = true
	return mrw.ResponseWriter.Write(b)
}

func (mrw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mrw.ResponseWriter
}

func (application *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsReceived.Add(1)

		mrw := newMetricsResponseWrite(w)
		next.ServeHTTP(w, r)

		totalResponsesSent.Add(1)
		totalResponsesSentByStatus.Add(strconv.Itoa(mrw.statusCode), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}

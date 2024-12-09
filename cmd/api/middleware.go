package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/thats-insane/awt-final/internal/data"
	"github.com/thats-insane/awt-final/internal/validator"
	"golang.org/x/time/rate"
)

func (a *appDependencies) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				w.Header().Set("Connection", "close")
				a.serverErr(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (a *appDependencies) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var mux sync.Mutex
	var clients = make(map[string]*client)

	go func() {
		for {
			time.Sleep(time.Minute)
			mux.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mux.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				a.serverErr(w, r, err)
				return
			}

			mux.Lock()
			_, found := clients[ip]
			if !found {
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(a.config.limiter.rps), a.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mux.Unlock()
				a.rateLimitExceed(w, r)
				return
			}

			mux.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

/* Authenticate a user using the token */
func (a *appDependencies) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = a.ctxSetUser(r, data.AnonUser)
			next.ServeHTTP(w, r)
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			a.invalidAuthToken(w, r)
			return
		}
		token := headerParts[1]
		v := validator.New()

		data.ValidateTokenPlaintext(v, token)
		if !v.IsEmpty() {
			a.invalidAuthToken(w, r)
			return
		}

		user, err := a.userModel.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				a.invalidAuthToken(w, r)
			default:
				a.serverErr(w, r, err)
			}
			return
		}

		r = a.ctxSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

/* Check if user is authenticated (not anonymous) */
func (a *appDependencies) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.ctxGetUser(r)

		if user.IsAnon() {
			a.authRequired(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

/* Check if user is activated */
func (a *appDependencies) requireActivated(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := a.ctxGetUser(r)

		if !user.Activated {
			a.inactiveAccount(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return a.requireAuth(fn)
}

func (a *appDependencies) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range a.config.cors.trustedOrigins {
				if origin == a.config.cors.trustedOrigins[i] {
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

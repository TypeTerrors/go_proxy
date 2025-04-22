package app

import (
	"net/http"
	"strings"
)

func (a *App) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != "" {
			a.Log.Info("new request:", "method", r.Method, "path", r.URL.Path, "host", r.Host)
		} else if r.Header.Get("X-Forwarded-Host") != "" {
			a.Log.Info("new request:", "method", r.Method, "path", r.URL.Path, "host", r.Header.Get("X-Forwarded-Host"))
		} else {
			a.Response(w, a.Err("host header and x-forwarded-host header not provided"), http.StatusNoContent)
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Host, a.name) {
			a.Response(w, a.Err("operation not permitted"), http.StatusServiceUnavailable)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.Response(w, a.Err("authorization header missing"), http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			a.Response(w, a.Err("invalid authorization header format"), http.StatusUnauthorized)
			return
		}
		token := parts[1]

		_, err := a.Jwt.ValidateJWT(token)
		if err != nil {
			a.Response(w, a.Err("authentication failed %s", err), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

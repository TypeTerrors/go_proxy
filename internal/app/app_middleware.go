package app

import (
	"net/http"
	"strings"
)

func (a *App) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Log.Info("new request:", "method", r.Method, "path", r.URL.Path, "host", r.Host)
		next.ServeHTTP(w, r)
	})
}

func (a *App) AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Host, a.name) {
			a.Response(w, a.Err("operation not permitted"), http.StatusServiceUnavailable)
		}

		_, err := a.Jwt.ValidateJWT(r.Header["Authorization"][0])
		if err != nil {
			a.Response(w, a.Err("authentication failed %s", err), http.StatusUnauthorized)
			return
		}
	})
}

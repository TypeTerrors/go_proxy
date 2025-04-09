package handlers

import "net/http"

func (a *App) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.Log.Info(r.Method, r.URL.Path)
		wrapped := &WrapperWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		a.Log.Info("New Request:", "method", r.Method, "status", wrapped.statusCode, "path", r.URL.Path, "host", r.Host)
		next.ServeHTTP(wrapped, r)
	})
}

type WrapperWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *WrapperWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

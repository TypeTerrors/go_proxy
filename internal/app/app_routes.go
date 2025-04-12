package app

import "net/http"

func (a *App) CreateRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", a.proxyRoutes())
	mux.Handle("/api/", a.apiRoutes())
	return a.LoggingMiddleware(mux)
}

func (a *App) proxyRoutes() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("GET /", a.HandleRequests)
    mux.HandleFunc("POST /", a.HandleRequests)
    return a.LoggingMiddleware(mux)
}

func (a *App) apiRoutes() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/add", a.HandleAddNewProxy)
    mux.HandleFunc("POST /api/del", a.HandleDeleteProxy)
    mux.HandleFunc("GET /api/tbl", a.HandleGetRedirectionRecords)
    return a.AuthenticationMiddleware(mux)
}


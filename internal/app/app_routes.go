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
	mux.HandleFunc("PUT /", a.HandleRequests)
	mux.HandleFunc("POST /", a.HandleRequests)
	mux.HandleFunc("PATCH /", a.HandleRequests)
	mux.HandleFunc("DELETE /", a.HandleRequests)
	return a.LoggingMiddleware(mux)
}

func (a *App) apiRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/prx", a.HandleGetRedirectionRecords)
	mux.HandleFunc("POST /api/prx", a.HandleAddNewProxy)
    mux.HandleFunc("PATCH /api/prx", a.HandlePatchProxy)
	mux.HandleFunc("DELETE /api/prx", a.HandleDeleteProxy)
	return a.AuthenticationMiddleware(mux)
}

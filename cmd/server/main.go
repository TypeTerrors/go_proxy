package main

import (
	"flag"
	"net/http"
	"prx/internal/handlers"
	"prx/internal/logger"
	"prx/internal/utils"
)

func main() {
	var redirectPairs utils.RedirectFlag

	flag.Var(&redirectPairs, "record", "Specify redirect records in the format FROM=TO. Multiple records can be specified.")
	flag.Parse()

	if err := utils.ParseRedirectRecords(redirectPairs); err != nil {
		logger.Log.Error("Failed to parse redirect records:", "error", err)
		return
	}

	router := http.NewServeMux()
	router.HandleFunc("GET /", handlers.HandleRequests)
	router.HandleFunc("POST /", handlers.HandleRequests)
	router.HandleFunc("GET /api/status", handlers.StatusHandler)

	middlewareStack := handlers.MiddlewareStack(handlers.LoggingMiddleware)

	server := &http.Server{
		Addr:    ":80",
		Handler: middlewareStack(router),
	}

	logger.Log.Info("Server started on port 80")
	if err := server.ListenAndServe(); err != nil {
		logger.Log.Fatal("Server failed to start:", "error", err)
	}
}
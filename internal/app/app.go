package app

import (
	"net/http"
	"os"
	"prx/internal/services"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type App struct {
	Jwt             *services.JWTService
	Log             *log.Logger
	Api             *http.Server
	Kube            services.Kube
	RedirectRecords map[string]string
	mu              sync.Mutex
	namespace       string
	name            string
}

func NewProxy(namespace, name, secret string, records map[string]string) *App {

	kube, err := services.NewKubeClient()
	if err != nil {
		panic(err)
	}

	app := &App{
		Jwt: services.NewJwtService(secret),
		Log: log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    true,
			ReportTimestamp: true,
			TimeFormat:      time.Kitchen,
			Prefix:          "go_proxy",
		}),
		RedirectRecords: records,
		Kube:            kube,
		namespace:       namespace,
		name:            name,
	}

	app.Api = &http.Server{
		Addr:    ":80",
		Handler: app.CreateRoutes(),
	}

	return app
}

func (a *App) Start() {

	a.Log.Info("Server started on port 80")
	if err := a.Api.ListenAndServe(); err != nil {
		a.Log.Fatal("Server failed to start:", "error", err)
	}
}

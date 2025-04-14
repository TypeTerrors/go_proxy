package app

import (
	"net/http"
	"os"
	"prx/internal/models"
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
	mu              sync.Mutex
	namespace       string
	name            string
	version         string
	RedirectRecords map[string]string
}

func NewProxy(settings models.NewProxySettings) *App {

	var err error

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "go_proxy",
	})

	app := &App{
		Jwt:             services.NewJwtService(settings.Secret),
		Log:             logger,
		RedirectRecords: settings.Records,
		namespace:       settings.Namespace,
		name:            settings.Name,
		version:         settings.Version,
	}

	app.Kube, err = services.NewKubeClient(logger)
	if err != nil {
		panic(err)
	}

	app.Api = &http.Server{
		Addr:    ":80",
		Handler: app.CreateRoutes(),
	}

	return app
}

func (a *App) Start() {

	jwt, err := a.Jwt.GenerateJWT()
	if err != nil {
		a.Log.Fatal("Server failed generate json web token on startup:", "error", err)
	}

	a.printSettings(jwt, os.Getenv("JWT_SECRET"))

	a.Log.Info("Server started on port 80")
	if err := a.Api.ListenAndServe(); err != nil {
		a.Log.Fatal("Server failed to start:", "error", err)
	}
}

package api

import (
	"net/http"
	"os"
	"prx/internal/models"
	"prx/internal/services"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

type Api struct {
	Jwt             *services.JWTService
	Log             *log.Logger
	Api             *http.Server
	Kube            services.Kube
	mu              sync.Mutex
	Namespace       string
	Name            string
	Version         string
	RedirectRecords map[string]string
}

func NewProxy(settings models.NewProxySettings) *Api {

	var err error

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.Kitchen,
		Prefix:          "go_proxy",
	})

	api := &Api{
		Jwt:             services.NewJwtService(settings.Secret),
		Log:             logger,
		RedirectRecords: settings.Records,
		Namespace:       settings.Namespace,
		Name:            settings.Name,
		Version:         settings.Version,
	}

	api.Kube, err = services.NewKubeClient(logger)
	if err != nil {
		panic(err)
	}

	api.Api = &http.Server{
		Addr:    ":80",
		Handler: api.CreateRoutes(),
	}

	return api
}

func (a *Api) Start(done chan<- error) {

	go func() {
		jwt, err := a.Jwt.GenerateJWT()
		if err != nil {
			a.Log.Fatal("Server failed generate json web token on startup:", "error", err)
		}

		a.printSettings(jwt, os.Getenv("JWT_SECRET"))

		a.Log.Info("Server started on port 80")
		if err := a.Api.ListenAndServe(); err != nil {
			done <- a.Err("failed to serve gRPCserver: %v", err)
		}
		done <- nil
	}()
}

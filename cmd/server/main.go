package main

import (
	"os"
	"prx/internal/app"
	"prx/internal/models"
)

var Version = "N/A"

func main() {

	prx := app.NewProxy(models.NewProxySettings{
		Name:      os.Getenv("NAMESPACE"),
		Namespace: os.Getenv("NAMESPACE"),
		Secret:    os.Getenv("JWT_SECRET"),
		Records:   make(map[string]string),
		Version:   Version,
	})

	prx.Start()
}

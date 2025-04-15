package main

import (
	"os"
	"prx/internal/api"
	"prx/internal/models"
	"prx/internal/rpc"
	"prx/internal/utils"
)

var Version = "N/A"

func main() {

	done := make(chan error, 2)

	prx := api.NewProxy(models.NewProxySettings{
		Name:      os.Getenv("NAMESPACE"),
		Namespace: os.Getenv("NAMESPACE"),
		Secret:    os.Getenv("JWT_SECRET"),
		Records:   make(map[string]string),
		Version:   Version,
	})

	rpc := rpc.NewGrpc(prx)

	prx.Start(done)
	rpc.Start(done)

	utils.GracefulShutdown(done)
}

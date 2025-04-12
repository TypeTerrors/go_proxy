package main

import (
	"os"
	"prx/internal/app"
)

func main() {

	prx := app.NewProxy(os.Getenv("NAMESPACE"), os.Getenv("NAMESPACE"), os.Getenv("JWT_SECRET"), make(map[string]string))
	prx.Start()
}

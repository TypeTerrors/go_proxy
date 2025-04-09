package main

import (
	"os"
	"prx/internal/handlers"
)

func main() {

	prx := handlers.NewProxy(os.Getenv("NAMESPACE"), os.Getenv("JWT_SECRET"), make(map[string]string))
	prx.Start()
}

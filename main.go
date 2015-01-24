package main

import (
	"os"
)

func main() {
	configFile := os.Getenv("IMGUR_GO_CONF")

	config := NewConfiguration(configFile)
	server := CreateServer(config)
	server.initServer()
}
